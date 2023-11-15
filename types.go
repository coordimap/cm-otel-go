package cmotel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// SpanAttrParentName span atrribute to mark the parent name
	SpanAttrParentName = "coordimap.span_attr.parent_name"

	// SpanAttrComponent span attribute to mark the component
	SpanAttrComponent = "coordimap.span_attr.component"

	// SpanAttrRelationship span attribute to mark a relationship
	SpanAttrRelationship = "coordimap.span_attr.relationship"

	// SpanAttrTargetService span attribute to mark a call or connection to another service. This means an outgoing relationship.
	SpanAttrTargetService = "coordimap.span_attr.target_service"
)

const (
	// EnvNodeNameType the environment variable that contains the node name
	EnvNodeNameType = "NODE_NAME"

	// EnvPodNameType the environment variable that contains the pod name
	EnvPodNameType = "POD_NAME"

	// PodNameCompleteType this is the name of the attribute that will hold the full internal name of the pod, <K8S_CLUSTER_NAME>.<NAMESPACE_NAME>.<POD_NAME>
	PodNameCompleteType = "POD_NAME"

	// EnvK8SNamespaceName the environment variable that contains the namespace name
	EnvK8SNamespaceName = "NAMESPACE_NAME"

	// EnvNodeIPType the environment variable that contains the node's primary IP
	EnvNodeIPType = "NODE_IP"

	// EnvServiceAccountType the environment variable that contains the node's service accout
	EnvServiceAccountType = "SERVICE_ACCOUNT"

	// EnvK8SClusterName the environment variable that contains the k8s cluster's name
	EnvK8SClusterName = "K8S_CLUSTER_NAME"

	// EnvServiceName environment variable that contains the name of the service
	EnvServiceName = "SERVICE_NAME"

	// EnvTracerName contains the name of the tracer to initialize the Coordimap Otel library
	EnvTracerName = "TRACER_NAME"

	// EnvCmPrefix contains the prefix of the related env variables
	EnvCmPrefix = "CM_PREFIX"

	// EnvServiceNamePrefix preferrably to be used in order to uniquely identify the services
	EnvServiceNamePrefix = "SERVICE_NAME_PREFIX"
)

const (
	// CmOtelComponentIDKey the key that will store the ID of the component
	CmOtelComponentIDKey = attribute.Key("cmotel.component.id")

	// CmOtelComponentInternalIDKey the key that will store the internal id of the component
	CmOtelComponentInternalIDKey = attribute.Key("cmotel.component.internal_id")
)

const (
	// TypeSpan marks the component as an otel span
	TypeSpan = "coordimap.otel.span"

	// ComponentType marks the component as a coordimap component
	ComponentType = "coordimap.cmotel.component_type"

	// CmotelComponentTypeKey Attribute key for the coordimap component type
	CmotelComponentTypeKey = attribute.Key("coordimap.cmotel.component_type")

	// ComponentTypeKey marks the component as a coordimap component
	ComponentTypeKey = attribute.Key(ComponentType)

	// ComponentRelationshipSkipInsert represents a relationship between two components
	ComponentRelationshipSkipInsert = "coordimap.cmotel.component_relationship_skipinsert"

	// OtelComponentRelationship represents a relationship between otel components
	OtelComponentRelationship = "coordimap.cmotel.component_relationship"

	// ComponentTypeGeneric The generic type for any component
	ComponentTypeGeneric = "coordimap.asset.generic"

	// ComponentTypeGenericContainer The generic container used by default
	ComponentTypeGenericContainer = "coordimap.asset.generic_container"
)

var (
	// CmOtelMessagingSystemNats sets the messaging system to NATS
	CmOtelMessagingSystemNats = semconv.MessagingSystemKey.String("NATS")
)

type contextKey string

// ContextKey is used to set or retrieve the cmOtel value to or from the context
const ContextKey contextKey = "cmotel"

type cmSpan struct {
	ctx  context.Context
	span trace.Span
}

type newSpanOpts struct {
	ctx          context.Context
	name         string
	parentName   string
	to           []string
	internalFrom []string
	externalFrom []string
}

type addComponentOpts struct {
	span          trace.Span
	componentType string
	isContainer   bool
	spanName      string
	attributes    []attribute.KeyValue
}

type addComponentOptionType = func(c *addComponentOpts) error

// SpanOption the function parameter for creating a Span
type SpanOption = func(c *newSpanOpts) error

type cmOtel struct {
	tracer             trace.Tracer
	serviceName        string
	spans              map[string]cmSpan
	relationships      map[string]string
	spanIDToNameMapper map[string]string
}

// CMOtel The interface that helps manage Coordimap spans
type CMOtel interface {
	NewSpan(opts ...SpanOption) (trace.Span, context.Context)
	EndSpan(name string, opts ...trace.SpanEndOption) error
	GetSpanContext(name string) (context.Context, error)
	SpanExists(name string) bool
	AddComponent(opts ...addComponentOptionType) error
	AddRemoteSpanCtx(spanCtx context.Context, spanName string) error
	GetSpanTraceparent(name string) string
	GetSpanTraceparentMaps(spanNames []string) (map[string]string, error)
	SetSpanFromTraceparent(name, traceparent string) error
}

// CMComponent describes the main values of the component
type CMComponent struct {
	Name       string            `json:"name"`
	InternalID string            `json:"internal_id"`
	Type       string            `json:"type"`
	Data       map[string]string `json:"data"`
}
