package evidence

import (
	"context"
	"testing"

	"github.com/AnouarMohamed/StateSight/internal/incidents"
	"github.com/AnouarMohamed/StateSight/internal/normalize"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func TestProvenanceAttributorRecordsCapturedGitAndKubectlSignals(t *testing.T) {
	analysis := analysisWithLiveResource(deploymentResource(2))
	attributions, err := (ProvenanceAttributor{}).BuildAttributions(context.Background(), analysis, replicaCandidate())
	if err != nil {
		t.Fatalf("build attributions: %v", err)
	}

	if len(attributions) != 2 {
		t.Fatalf("expected Git and live provenance records, got %d", len(attributions))
	}
	if attributions[0].Source != "git" || attributions[0].Actor != unattributedActor || attributions[0].Confidence != 1 {
		t.Fatalf("unexpected Git attribution: %#v", attributions[0])
	}
	if attributions[0].Metadata["revision"] != analysis.DesiredRevision || attributions[0].Metadata["repository"] != analysis.Source.RepoURL {
		t.Fatalf("expected Git provenance metadata, got %#v", attributions[0].Metadata)
	}
	if attributions[1].Source != "kubectl" || attributions[1].Actor != unattributedActor || attributions[1].Confidence != 1 {
		t.Fatalf("unexpected live attribution: %#v", attributions[1])
	}
	if trusted, ok := attributions[1].Metadata["trusted_observation"].(bool); !ok || !trusted {
		t.Fatalf("expected trusted kubectl observation, got %#v", attributions[1].Metadata)
	}
}

func TestProvenanceAttributorUsesLatestExactManagedFieldsOwner(t *testing.T) {
	live := deploymentResource(2)
	metadata := live["metadata"].(map[string]any)
	metadata["managedFields"] = []any{
		managedReplicasEntry("deployment-controller", "2026-05-25T12:00:00Z"),
		managedReplicasEntry("horizontal-pod-autoscaler", "2026-05-25T13:00:00Z"),
	}

	attributions, err := (ProvenanceAttributor{}).BuildAttributions(
		context.Background(),
		analysisWithLiveResource(live),
		replicaCandidate(),
	)
	if err != nil {
		t.Fatalf("build attributions: %v", err)
	}

	if len(attributions) != 3 {
		t.Fatalf("expected managedFields evidence in addition to provenance, got %d", len(attributions))
	}
	managed := attributions[2]
	if managed.Source != "managedFields" || managed.Actor != "horizontal-pod-autoscaler" {
		t.Fatalf("unexpected managedFields attribution: %#v", managed)
	}
	if managed.Metadata["matching_entries"] != 2 {
		t.Fatalf("expected matching ownership count in metadata, got %#v", managed.Metadata)
	}
}

func TestProvenanceAttributorMatchesContainerImageOwnershipByContainerName(t *testing.T) {
	live := deploymentResource(2)
	live["spec"].(map[string]any)["template"] = map[string]any{
		"spec": map[string]any{
			"containers": []any{map[string]any{"name": "ledger-api", "image": "registry.example.com/ledger-api:live"}},
		},
	}
	live["metadata"].(map[string]any)["managedFields"] = []any{map[string]any{
		"manager":    "argocd-controller",
		"operation":  "Apply",
		"apiVersion": "apps/v1",
		"time":       "2026-05-25T13:00:00Z",
		"fieldsV1": map[string]any{
			"f:spec": map[string]any{
				"f:template": map[string]any{
					"f:spec": map[string]any{
						"f:containers": map[string]any{
							`k:{"name":"ledger-api"}`: map[string]any{
								"f:image": map[string]any{},
							},
						},
					},
				},
			},
		},
	}}
	candidate := replicaCandidate()
	candidate.FieldPath = "spec.template.spec.containers[0].image"

	attributions, err := (ProvenanceAttributor{}).BuildAttributions(
		context.Background(),
		analysisWithLiveResource(live),
		candidate,
	)
	if err != nil {
		t.Fatalf("build attributions: %v", err)
	}
	if len(attributions) != 3 || attributions[2].Actor != "argocd-controller" {
		t.Fatalf("expected exact container-image manager evidence, got %#v", attributions)
	}
}

func TestProvenanceAttributorMatchesMetadataLabelOwnership(t *testing.T) {
	live := deploymentResource(2)
	live["metadata"].(map[string]any)["labels"] = map[string]any{"app.kubernetes.io/name": "ledger-api-live"}
	live["metadata"].(map[string]any)["managedFields"] = []any{map[string]any{
		"manager":    "argocd-controller",
		"operation":  "Apply",
		"apiVersion": "apps/v1",
		"time":       "2026-05-25T13:00:00Z",
		"fieldsV1": map[string]any{
			"f:metadata": map[string]any{
				"f:labels": map[string]any{
					"f:app.kubernetes.io/name": map[string]any{},
				},
			},
		},
	}}
	candidate := replicaCandidate()
	candidate.FieldPath = "metadata.labels.app.kubernetes.io/name"

	attributions, err := (ProvenanceAttributor{}).BuildAttributions(
		context.Background(),
		analysisWithLiveResource(live),
		candidate,
	)
	if err != nil {
		t.Fatalf("build attributions: %v", err)
	}
	if len(attributions) != 3 || attributions[2].Actor != "argocd-controller" {
		t.Fatalf("expected exact metadata-label manager evidence, got %#v", attributions)
	}
}

func TestProvenanceAttributorAttributesOnlyExactNamedContainerResourceOwnership(t *testing.T) {
	live := deploymentResource(2)
	live["spec"].(map[string]any)["template"] = map[string]any{
		"spec": map[string]any{
			"containers": []any{map[string]any{
				"name": "ledger-api",
				"env":  []any{map[string]any{"name": "MODE", "value": "live"}},
				"resources": map[string]any{
					"requests": map[string]any{"nvidia.com/gpu": "1"},
				},
			}},
		},
	}
	live["metadata"].(map[string]any)["managedFields"] = []any{map[string]any{
		"manager":    "rollout-controller",
		"operation":  "Apply",
		"apiVersion": "apps/v1",
		"time":       "2026-05-25T13:00:00Z",
		"fieldsV1": map[string]any{
			"f:spec": map[string]any{
				"f:template": map[string]any{
					"f:spec": map[string]any{
						"f:containers": map[string]any{
							`k:{"name":"ledger-api"}`: map[string]any{
								"f:env": map[string]any{
									`k:{"name":"MODE"}`: map[string]any{},
								},
								"f:resources": map[string]any{
									"f:requests": map[string]any{
										"f:nvidia.com/gpu": map[string]any{},
									},
								},
							},
						},
					},
				},
			},
		},
	}}

	resourceCandidate := replicaCandidate()
	resourceCandidate.FieldPath = "spec.template.spec.containers[name=ledger-api].resources.requests.nvidia.com/gpu"
	resourceAttributions, err := (ProvenanceAttributor{}).BuildAttributions(
		context.Background(),
		analysisWithLiveResource(live),
		resourceCandidate,
	)
	if err != nil {
		t.Fatalf("build resource attributions: %v", err)
	}
	if len(resourceAttributions) != 3 || resourceAttributions[2].Actor != "rollout-controller" {
		t.Fatalf("expected exact named-container resource evidence, got %#v", resourceAttributions)
	}

	envCandidate := replicaCandidate()
	envCandidate.FieldPath = "spec.template.spec.containers[name=ledger-api].env[name=MODE]"
	envAttributions, err := (ProvenanceAttributor{}).BuildAttributions(
		context.Background(),
		analysisWithLiveResource(live),
		envCandidate,
	)
	if err != nil {
		t.Fatalf("build environment attributions: %v", err)
	}
	if len(envAttributions) != 2 {
		t.Fatalf("expected aggregate environment drift not to claim partial manager ownership, got %#v", envAttributions)
	}
}

func TestProvenanceAttributorMatchesQualifiedServiceSelectorOwnership(t *testing.T) {
	desired := serviceSelectorResource("ledger-api")
	live := serviceSelectorResource("other-service")
	live["metadata"].(map[string]any)["managedFields"] = []any{map[string]any{
		"manager":    "service-operator",
		"operation":  "Update",
		"apiVersion": "v1",
		"time":       "2026-05-25T13:00:00Z",
		"fieldsV1": map[string]any{
			"f:spec": map[string]any{
				"f:selector": map[string]any{
					"f:app.kubernetes.io/name": map[string]any{},
				},
			},
		},
	}}
	candidate := incidents.Candidate{
		ResourceKey:  "Service|payments|ledger-api",
		ResourceRef:  "v1/Service:payments/ledger-api",
		FieldPath:    "spec.selector.app.kubernetes.io/name",
		DesiredValue: "ledger-api",
		LiveValue:    "other-service",
	}

	attributions, err := (ProvenanceAttributor{}).BuildAttributions(
		context.Background(),
		analysisWithResources(desired, live),
		candidate,
	)
	if err != nil {
		t.Fatalf("build attributions: %v", err)
	}
	if len(attributions) != 3 || attributions[2].Actor != "service-operator" {
		t.Fatalf("expected exact service-selector manager evidence, got %#v", attributions)
	}
}

func TestProvenanceAttributorDoesNotUseManagedFieldsForAnotherPath(t *testing.T) {
	live := deploymentResource(2)
	live["metadata"].(map[string]any)["managedFields"] = []any{
		managedReplicasEntry("horizontal-pod-autoscaler", "2026-05-25T13:00:00Z"),
	}
	candidate := replicaCandidate()
	candidate.FieldPath = "metadata.annotations.example.com/owner"

	attributions, err := (ProvenanceAttributor{}).BuildAttributions(
		context.Background(),
		analysisWithLiveResource(live),
		candidate,
	)
	if err != nil {
		t.Fatalf("build attributions: %v", err)
	}
	if len(attributions) != 2 {
		t.Fatalf("expected no unmatched managedFields attribution, got %#v", attributions)
	}
}

func TestProvenanceAttributorMarksSyntheticStateAsUntrusted(t *testing.T) {
	live := deploymentResource(2)
	live["metadata"].(map[string]any)["managedFields"] = []any{
		managedReplicasEntry("invented-manager", "2026-05-25T13:00:00Z"),
	}
	analysis := analysisWithLiveResource(live)
	analysis.LiveSummary = map[string]any{"source": "synthetic"}

	attributions, err := (ProvenanceAttributor{}).BuildAttributions(context.Background(), analysis, replicaCandidate())
	if err != nil {
		t.Fatalf("build attributions: %v", err)
	}

	if len(attributions) != 2 {
		t.Fatalf("expected synthetic state not to emit managedFields attribution, got %#v", attributions)
	}
	liveAttribution := attributions[1]
	if liveAttribution.Source != "synthetic" || liveAttribution.Confidence != 0 {
		t.Fatalf("unexpected synthetic attribution: %#v", liveAttribution)
	}
	if trusted, ok := liveAttribution.Metadata["trusted_observation"].(bool); !ok || trusted {
		t.Fatalf("expected synthetic signal to be untrusted, got %#v", liveAttribution.Metadata)
	}
}

func analysisWithLiveResource(live map[string]any) AnalysisContext {
	return analysisWithResources(deploymentResource(3), live)
}

func analysisWithResources(desired, live map[string]any) AnalysisContext {
	normalizer := normalize.PassThroughNormalizer{}
	return AnalysisContext{
		Application: model.Application{Name: "ledger-api", Namespace: "payments"},
		Source: model.SourceDefinition{
			RepoURL:       "https://github.com/example/platform-config",
			DefaultBranch: "main",
			Path:          "clusters/prod",
		},
		Cluster:         model.Cluster{ID: "cluster-1", Name: "prod-eu"},
		DesiredRevision: "abcdef1234567890",
		Desired:         normalizer.Normalize([]map[string]any{desired}),
		Live:            normalizer.Normalize([]map[string]any{live}),
		LiveSummary:     map[string]any{"source": "kubectl"},
	}
}

func replicaCandidate() incidents.Candidate {
	return incidents.Candidate{
		ResourceKey:  "Deployment|payments|ledger-api",
		ResourceRef:  "apps/v1/Deployment:payments/ledger-api",
		FieldPath:    "spec.replicas",
		DesiredValue: "3",
		LiveValue:    "2",
	}
}

func deploymentResource(replicas int) map[string]any {
	return map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			"name":      "ledger-api",
			"namespace": "payments",
		},
		"spec": map[string]any{
			"replicas": replicas,
		},
	}
}

func serviceSelectorResource(selector string) map[string]any {
	return map[string]any{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]any{
			"name":      "ledger-api",
			"namespace": "payments",
		},
		"spec": map[string]any{
			"selector": map[string]any{"app.kubernetes.io/name": selector},
		},
	}
}

func managedReplicasEntry(manager, observedAt string) map[string]any {
	return map[string]any{
		"manager":    manager,
		"operation":  "Update",
		"apiVersion": "apps/v1",
		"time":       observedAt,
		"fieldsV1": map[string]any{
			"f:spec": map[string]any{
				"f:replicas": map[string]any{},
			},
		},
	}
}
