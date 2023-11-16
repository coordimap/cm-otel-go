package cmotel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// WithSpanInternalRelationshipFrom the span names that are related, i.e. come after, this span.
func WithSpanInternalRelationshipFrom(from string) SpanOption {
	return func(opt *newSpanOpts) error {
		opt.internalFrom = append(opt.internalFrom, from)
		return nil
	}
}

// WithSpanExternalRelationshipFrom the span names that are related, i.e. come after, this span.
func WithSpanExternalRelationshipFrom(from string) SpanOption {
	return func(opt *newSpanOpts) error {
		opt.externalFrom = append(opt.externalFrom, from)
		return nil
	}
}

// WithSpanRelationshipTo currently not implemented
func WithSpanRelationshipTo(to string) SpanOption {
	return func(opt *newSpanOpts) error {
		return nil
	}
}

// WithSpanContext an existing span context to create the span from. It defaults to Background()
func WithSpanContext(ctx context.Context) SpanOption {
	return func(opt *newSpanOpts) error {
		opt.ctx = ctx

		return nil
	}
}

// WithSpanName the name of the span. It must be a non empty string
func WithSpanName(name string) SpanOption {
	return func(opt *newSpanOpts) error {
		if name == "" {
			return errors.New("name must not be empty")
		} else if strings.Contains(name, "@") {
			return errors.New("name must not contain @")
		}

		opt.name = name

		return nil
	}
}

// WithParentSpanName provide the name of the parent span. The name must exist otherwise it will return an error.
func WithParentSpanName(name string) SpanOption {
	return func(opt *newSpanOpts) error {
		if name == "" {
			return errors.New("parent name must not be empty")
		}

		opt.parentName = name

		return nil
	}
}

func (cm *cmOtel) NewSpan(opts ...SpanOption) (trace.Span, context.Context) {
	spanOpts := &newSpanOpts{
		ctx:          context.Background(),
		name:         "",
		parentName:   "",
		to:           []string{},
		internalFrom: []string{},
		externalFrom: []string{},
	}
	newSpanOpts := []trace.SpanStartOption{}

	for _, opt := range opts {
		opt(spanOpts)
	}

	hasParentConfig := false
	parentSpanID := ""

	if spanOpts.parentName != "" {
		if !cm.SpanExists(spanOpts.parentName) {
			return nil, context.TODO()
		}

		spanOpts.ctx = cm.spans[spanOpts.parentName].ctx
		parentSpanID = cm.spans[spanOpts.parentName].span.SpanContext().SpanID().String()
		hasParentConfig = true
	}

	if spanOpts.ctx != context.Background() && spanOpts.ctx != context.TODO() {
		link := trace.LinkFromContext(spanOpts.ctx)

		if link.SpanContext.SpanID().IsValid() {
			parentSpanID = link.SpanContext.SpanID().String()
			parentName, ok := cm.spanIDToNameMapper[parentSpanID]

			if ok {
				spanOpts.parentName = parentName
				hasParentConfig = true
			}
		}
	}

	if hasParentConfig {
		// needed to generate the respective relationship
		newSpanOpts = append(newSpanOpts, trace.WithAttributes(attribute.KeyValue{
			Key:   SpanAttrParentName,
			Value: attribute.StringValue(cm.generateInternalName(spanOpts.parentName)),
		}))
	}

	spanLinks := []trace.Link{}

	for _, internalFrom := range spanOpts.internalFrom {
		spanLinks = append(spanLinks, trace.Link{
			SpanContext: trace.SpanContextFromContext(cm.spans[internalFrom].ctx),
			Attributes: []attribute.KeyValue{
				attribute.String(SpanAttrRelationship, fmt.Sprintf("%s@@@%s", cm.generateInternalName(internalFrom), cm.generateInternalName(spanOpts.name))),
			},
		})
	}

	for _, from := range spanOpts.externalFrom {
		spanLinks = append(spanLinks, trace.Link{
			SpanContext: trace.SpanContextFromContext(cm.spans[from].ctx),
			Attributes: []attribute.KeyValue{
				attribute.String(SpanAttrRelationship, fmt.Sprintf("%s@@@%s", from, cm.generateInternalName(spanOpts.name))),
			},
		})
	}

	if len(spanLinks) != 0 {
		newSpanOpts = append(newSpanOpts, trace.WithLinks(spanLinks...))
	}

	ctx, span := cm.tracer.Start(
		spanOpts.ctx,
		cm.generateInternalName(spanOpts.name),
		newSpanOpts...,
	)

	cm.spans[spanOpts.name] = cmSpan{
		ctx:  ctx,
		span: span,
	}

	cm.spanIDToNameMapper[cm.spans[spanOpts.name].span.SpanContext().SpanID().String()] = spanOpts.name

	return span, ctx
}

func (cm *cmOtel) SpanExists(name string) bool {
	if _, ok := cm.spans[name]; ok {
		return true
	}

	return false
}

func (cm *cmOtel) EndSpan(name string, opts ...trace.SpanEndOption) error {
	span, ok := cm.spans[name]
	if !ok {
		return fmt.Errorf("span %s does not exist", name)
	}

	span.span.End(opts...)

	return nil
}

func (cm *cmOtel) GetSpanContext(name string) (context.Context, error) {
	span, ok := cm.spans[name]
	if !ok {
		return context.TODO(), fmt.Errorf("span %s does not exist", name)
	}

	return span.ctx, nil
}

// WithAddComponentSpan the Span where to add the coordimap component annotation to
func WithAddComponentSpan(span trace.Span) addComponentOptionType {
	return func(opt *addComponentOpts) error {
		opt.span = span

		return nil
	}
}

// WithAddComponentSpanName the Span name where to add the coordimap component annotation to
func WithAddComponentSpanName(spanName string) addComponentOptionType {
	return func(opt *addComponentOpts) error {
		opt.spanName = spanName

		return nil
	}
}

// WithAddComponentType the type of the coordimap component
func WithAddComponentType(componentType string) addComponentOptionType {
	return func(opt *addComponentOpts) error {
		if componentType == "" {
			return errors.New("component type must not be empty")
		}
		opt.componentType = componentType

		return nil
	}
}

// WithAddComponentAttribute extra attributes to add to the component
func WithAddComponentAttribute(attribute attribute.KeyValue) addComponentOptionType {
	return func(opt *addComponentOpts) error {
		opt.attributes = append(opt.attributes, attribute)

		return nil
	}
}

func (cm *cmOtel) AddComponent(opts ...addComponentOptionType) error {
	options := &addComponentOpts{
		span:          nil,
		componentType: "",
		spanName:      "",
		attributes:    []attribute.KeyValue{},
		isContainer:   false,
	}

	for _, opt := range opts {
		if err := opt(options); err != nil {
			return err
		}
	}

	if options.span == nil && options.spanName == "" {
		return errors.New("either span or the span name must be provided")
	}

	// if span is not set but the spanName is then try to retrieve it
	if options.span == nil && options.spanName != "" {
		if !cm.SpanExists(options.spanName) {
			return fmt.Errorf("span %s does not exist", options.spanName)
		}

		options.span = cm.spans[options.spanName].span
	}

	// check if span is set then get the span name from it's ID
	if options.span != nil {
		spanID := options.span.SpanContext().SpanID().String()
		spanName, ok := cm.spanIDToNameMapper[spanID]
		if !ok {
			return fmt.Errorf("the provided span with spanID %s does not exist", spanID)
		}

		options.spanName = spanName
	}

	if options.componentType != "" && !cm.SpanExists(options.spanName) {
		return fmt.Errorf("span %s does not exist", options.spanName)
	}

	newComponentData := map[string]string{}
	for _, attr := range options.attributes {
		newComponentData[string(attr.Key)] = attr.Value.AsString()
	}

	newComponent := CMComponent{
		InternalID: cm.generateInternalName(options.spanName),
		Name:       options.spanName,
		Type:       options.componentType,
		Data:       newComponentData,
	}

	marshaledNewComponent, errMarshaledNewComponent := json.Marshal(newComponent)
	if errMarshaledNewComponent != nil {
		return errors.Join(errors.New("cannot marshal the component"), errMarshaledNewComponent)
	}

	options.span.SetAttributes([]attribute.KeyValue{
		{
			Key:   SpanAttrComponent,
			Value: attribute.StringValue(string(marshaledNewComponent)),
		},
	}...)

	return nil
}

func (cm *cmOtel) generateInternalName(name string) string {
	if strings.Contains(name, "@") {
		return name
	}

	return fmt.Sprintf("%s@%s", GetServiceName(cm.serviceName), name)
}

// AddRemoteSpanCtx load the remote span context and name in the library in order to use it for relationships
func (cm *cmOtel) AddRemoteSpanCtx(spanCtx context.Context, spanName string) error {
	if _, exists := cm.spans[spanName]; exists {
		return fmt.Errorf("span %s already exists", spanName)
	}

	cm.spans[spanName] = cmSpan{
		ctx:  spanCtx,
		span: trace.SpanFromContext(spanCtx),
	}

	return nil
}

// GetSpanTraceparent returns the traceparent string for an existing span
func (cm *cmOtel) GetSpanTraceparent(name string) string {
	if !cm.SpanExists(name) {
		return ""
	}

	spanCtx := cm.spans[name].span.SpanContext()

	return fmt.Sprintf("00-%s-%s-%s", spanCtx.TraceID().String(), spanCtx.SpanID().String(), spanCtx.TraceFlags().String())
}

// GetSpanAsHeader returns the traceparent string for an existing span
func (cm *cmOtel) GetSpanTraceparentMaps(spanNames []string) (map[string]string, error) {
	allSpans := map[string]string{}

	for _, name := range spanNames {
		if !cm.SpanExists(name) {
			return map[string]string{}, fmt.Errorf("span %s does not exist", name)
		}

		spanCtx := cm.spans[name].span.SpanContext()

		allSpans[cm.generateInternalName(name)] = fmt.Sprintf("00-%s-%s-%s", spanCtx.TraceID().String(), spanCtx.SpanID().String(), spanCtx.TraceFlags().String())
	}

	return allSpans, nil
}

// GetSpanTraceparent returns the traceparent string for an existing span
func (cm *cmOtel) SetSpanFromTraceparent(name, traceparent string) error {
	if cm.SpanExists(name) {
		return nil
	}

	ctx, errParseTraceparent := ParseTraceParent(traceparent)
	if errParseTraceparent != nil {
		return errors.Join(errors.New("could not parse the provided traceparent"), errParseTraceparent)
	}

	if errAddRemoteSpan := cm.AddRemoteSpanCtx(ctx, name); errAddRemoteSpan != nil {
		return errors.Join(fmt.Errorf("could not add remote span with name %s", name), errAddRemoteSpan)
	}

	cm.spanIDToNameMapper[trace.SpanContextFromContext(cm.spans[name].ctx).SpanID().String()] = name

	return nil
}
