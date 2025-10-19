# SCAI v2: Technical Considerations & Pitfalls to Avoid

## Critical Technical Recommendations

### 1. Start Small: Minimal Viable Control Plane

**DON'T:**
```
❌ Build all 6 phases before testing
❌ Perfect RAG system before proving operator works
❌ Complex multi-cloud from day 1
```

**DO:**
```
✅ Phase 1: Basic operator (AWS EC2 only)
✅ Test with 3 real deployments
✅ Get feedback, iterate
✅ Add RAG/Crossplane/GraphQL incrementally
```

**Why:** Kubernetes operators are complex. Prove the concept works before investing months.

---

### 2. Operator Development: Lessons from Production Systems

#### A. Use Kubebuilder, Not Hand-Rolled Code

**DON'T:**
```go
// ❌ Manual controller implementation
func main() {
    clientset := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
    for {
        deployments, _ := clientset.AppsV1().Deployments("").List(...)
        // Manual reconciliation logic
        time.Sleep(5 * time.Second)
    }
}
```

**DO:**
```go
// ✅ Kubebuilder scaffolding
import ctrl "sigs.k8s.io/controller-runtime"

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Framework handles watches, retries, queuing
}
```

**Why:** Kubebuilder provides:
- Leader election (for HA)
- Event filtering (reduce CPU usage)
- Automatic retries (with exponential backoff)
- Prometheus metrics (out-of-box)

**Tool:** `kubebuilder init --domain scai.io --repo github.com/smana/scai`

#### B. Idempotency is Critical

**Problem:** Reconcile loop runs every 5 minutes. If non-idempotent, creates duplicate resources.

**BAD:**
```go
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ❌ Always creates new EC2 instance
    instanceID, err := r.AWS.CreateEC2Instance(...)
    deployment.Status.InstanceID = instanceID
    return ctrl.Result{}, nil
}
```

**GOOD:**
```go
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    deployment := &scaiapi.Deployment{}
    r.Get(ctx, req.NamespacedName, deployment)

    // ✅ Check if infrastructure already exists
    if deployment.Status.InfrastructureRef != nil {
        infra := &scaiapi.Infrastructure{}
        r.Get(ctx, deployment.Status.InfrastructureRef, infra)
        if infra.Status.Phase == "Ready" {
            return ctrl.Result{}, nil  // Already reconciled
        }
    }

    // Create infrastructure only if missing
    infra := r.buildInfrastructure(deployment)
    r.Create(ctx, infra)
    deployment.Status.InfrastructureRef = infra.ObjectRef()
    r.Status().Update(ctx, deployment)
    return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}
```

**Test:**
```bash
# Apply same Deployment CR 3 times
kubectl apply -f deployment.yaml
kubectl apply -f deployment.yaml
kubectl apply -f deployment.yaml

# Verify: Only 1 Infrastructure CR created (not 3)
kubectl get infrastructure
```

#### C. Handle Finalizers for Cleanup

**Problem:** When user deletes Deployment CR, cloud resources remain orphaned.

**Solution:** Add finalizer to ensure cleanup runs.

```go
const deploymentFinalizer = "scai.io/finalizer"

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    deployment := &scaiapi.Deployment{}
    if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Check if deployment is being deleted
    if !deployment.DeletionTimestamp.IsZero() {
        if controllerutil.ContainsFinalizer(deployment, deploymentFinalizer) {
            // Delete cloud resources
            if err := r.deleteInfrastructure(ctx, deployment); err != nil {
                return ctrl.Result{}, err
            }

            // Remove finalizer
            controllerutil.RemoveFinalizer(deployment, deploymentFinalizer)
            if err := r.Update(ctx, deployment); err != nil {
                return ctrl.Result{}, err
            }
        }
        return ctrl.Result{}, nil
    }

    // Add finalizer if not present
    if !controllerutil.ContainsFinalizer(deployment, deploymentFinalizer) {
        controllerutil.AddFinalizer(deployment, deploymentFinalizer)
        if err := r.Update(ctx, deployment); err != nil {
            return ctrl.Result{}, err
        }
    }

    // Normal reconciliation...
}
```

#### D. Status Updates: Use Sub-Resource

**Problem:** Updating `.spec` and `.status` in same update causes conflicts.

**BAD:**
```go
// ❌ Updates entire object (spec + status)
deployment.Status.Phase = "Running"
r.Update(ctx, deployment)  // Race condition with spec changes
```

**GOOD:**
```go
// ✅ Update status sub-resource separately
deployment.Status.Phase = "Running"
r.Status().Update(ctx, deployment)  // No conflict with spec
```

**Why:** Kubernetes separates `.spec` (user intent) and `.status` (observed state). Status updates use different endpoint to avoid conflicts.

---

### 3. RAG System: Performance & Accuracy

#### A. Embedding Strategy: Chunk Size Matters

**Problem:** Embedding entire deployment metadata (10KB+) loses semantic meaning.

**BAD:**
```python
# ❌ Embed everything (too much noise)
text = f"""
Deployment: {deployment.name}
Namespace: {deployment.namespace}
Framework: {deployment.spec.framework}
...
(3000 lines of JSON metadata)
"""
embedding = model.encode(text)
```

**GOOD:**
```python
# ✅ Embed only semantically relevant fields
text = f"""
Framework: {deployment.spec.framework}
Language: {analysis.language}
User Intent: {deployment.spec.prompt}
Key Dependencies: {', '.join(analysis.dependencies[:10])}
Deployment Strategy: {deployment.status.strategy}
Outcome: {deployment.status.outcome}
"""
embedding = model.encode(text)
```

**Why:** Embeddings capture semantic similarity. Too much text = diluted meaning.

**Rule:** Keep embedding text < 500 words.

#### B. Vector Index Tuning (pgvector)

**Problem:** Linear scan of 100k vectors = slow queries (>5 seconds).

**Solution:** Use IVFFlat index with appropriate `lists` parameter.

```sql
-- ❌ Default: No index (slow for >10k vectors)
CREATE TABLE deployment_embeddings (
    embedding vector(384)
);

-- ✅ IVFFlat index (100x faster)
CREATE INDEX ON deployment_embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);  -- sqrt(total_rows) is good default

-- For 100k vectors: lists = sqrt(100000) ≈ 316
-- For 10k vectors:  lists = sqrt(10000) ≈ 100
```

**Benchmark:**
```python
import time

# Measure query time
start = time.time()
results = cursor.execute(
    "SELECT * FROM deployment_embeddings ORDER BY embedding <=> %s LIMIT 5",
    (query_embedding,)
)
end = time.time()

print(f"Query time: {end - start:.2f}s")
# Target: <100ms for 100k vectors
```

**If still slow:**
- Reduce embedding dimensions (384 → 256)
- Use HNSW index (better accuracy, more memory)
- Migrate to Weaviate (10x faster than pgvector)

#### C. Re-Ranking for Better Accuracy

**Problem:** Vector similarity alone misses important filters (budget, region).

**Solution:** Two-stage retrieval.

```python
def retrieve_with_reranking(query_embedding, filters, top_k=5):
    # Stage 1: Vector similarity (retrieve 20 candidates)
    cursor.execute(
        """
        SELECT *,
            1 - (embedding <=> %s) AS similarity
        FROM deployment_embeddings
        WHERE framework = %s  -- Pre-filter by framework
          AND cost_monthly <= %s  -- Pre-filter by budget
        ORDER BY embedding <=> %s
        LIMIT 20  -- Retrieve more than needed
        """,
        (query_embedding, filters['framework'], filters['max_cost'], query_embedding)
    )
    candidates = cursor.fetchall()

    # Stage 2: Re-rank by business logic
    for candidate in candidates:
        score = (
            0.5 * candidate['similarity'] +
            0.2 * recency_score(candidate['created_at']) +
            0.2 * feedback_score(candidate['user_feedback']) +
            0.1 * cost_efficiency_score(candidate['cost_monthly'])
        )
        candidate['final_score'] = score

    # Return top K after re-ranking
    return sorted(candidates, key=lambda x: x['final_score'], reverse=True)[:top_k]
```

**Why:** Hybrid search (vector + filters + business logic) beats pure vector similarity.

---

### 4. Crossplane: Composition Best Practices

#### A. Use KCL for Logic, Not YAML

**Problem:** YAML compositions become unreadable with conditionals.

**BAD (YAML with Patches):**
```yaml
# ❌ Unreadable: 200 lines of patches
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
spec:
  patchSets:
    - name: instance-type
      patches:
        - type: FromCompositeFieldPath
          fromFieldPath: spec.size
          toFieldPath: spec.forProvider.instanceType
          transforms:
            - type: map
              map:
                small: t3.small
                medium: t3.medium
                large: t3.large
```

**GOOD (KCL):**
```python
# ✅ Readable: Logic in code
import crossplane.v1 as cp

schema VMSpec:
    size: str  # small, medium, large

def instance_type(size: str) -> str:
    {
        "small": "t3.small",
        "medium": "t3.medium",
        "large": "t3.large",
    }[size]

composition VMDeployment:
    spec: VMSpec
    instance = instance_type(spec.size)

    resources: [
        cp.ComposedResource {
            base: ec2.Instance {
                spec.forProvider.instanceType: instance
            }
        }
    ]
```

**Why:** KCL is Turing-complete (loops, functions), YAML is not.

#### B. Test Compositions Locally

**Problem:** Deploy composition to cluster → fails → debug → repeat (slow cycle).

**Solution:** Use Crossplane CLI to render compositions locally.

```bash
# Install Crossplane CLI
curl -sL https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh | sh

# Render composition locally (no cluster needed)
crossplane beta render \
  deployment.yaml \
  composition.yaml \
  function-kcl.yaml \
  --output rendered.yaml

# Verify: Check rendered resources
cat rendered.yaml
```

**Why:** Fast feedback loop (seconds vs minutes).

#### C. Versioning Compositions

**Problem:** Update composition → breaks existing deployments.

**Solution:** Use `compositionUpdatePolicy: Manual`.

```yaml
apiVersion: scai.io/v1alpha1
kind: Infrastructure
spec:
  compositionRef:
    name: scai-vm-deployment
  compositionUpdatePolicy: Manual  # Don't auto-update
```

**Upgrade Process:**
1. Create new composition version: `scai-vm-deployment-v2`
2. Test with new deployments
3. Gradually migrate old deployments: `kubectl patch infrastructure X --type=merge -p '{"spec":{"compositionRef":{"name":"scai-vm-deployment-v2"}}}'`

**Why:** Prevents accidental breaking changes.

---

### 5. GraphQL API: Performance Pitfalls

#### A. N+1 Query Problem

**Problem:** Fetching 100 deployments → 100 database queries.

**BAD:**
```go
// ❌ N+1 queries
func (r *queryResolver) Deployments(ctx context.Context) ([]*model.Deployment, error) {
    deployments := []*scaiapi.Deployment{}
    r.k8sClient.List(ctx, deployments)

    result := []*model.Deployment{}
    for _, deploy := range deployments {
        // Fetch infrastructure for each deployment (N queries)
        infra := &scaiapi.Infrastructure{}
        r.k8sClient.Get(ctx, deploy.Status.InfrastructureRef, infra)
        result = append(result, convertToGraphQL(deploy, infra))
    }
    return result
}
```

**GOOD (Use DataLoader):**
```go
// ✅ Batched queries
import "github.com/graph-gophers/dataloader"

func (r *queryResolver) Deployments(ctx context.Context) ([]*model.Deployment, error) {
    deployments := []*scaiapi.Deployment{}
    r.k8sClient.List(ctx, deployments)

    // Batch-load all infrastructures in one query
    infraIDs := []string{}
    for _, deploy := range deployments {
        infraIDs = append(infraIDs, deploy.Status.InfrastructureRef.Name)
    }
    infras := r.InfrastructureLoader.LoadMany(ctx, infraIDs)

    result := []*model.Deployment{}
    for i, deploy := range deployments {
        result = append(result, convertToGraphQL(deploy, infras[i]))
    }
    return result
}
```

**Why:** 1 query vs 100 queries = 50x faster.

#### B. Query Complexity Limits

**Problem:** Malicious user sends deeply nested query, crashes server.

```graphql
# ❌ 10-level deep query (exponential DB queries)
{
  deployments {
    infrastructure {
      resources {
        dependencies {
          references {
            # ... 5 more levels
          }
        }
      }
    }
  }
}
```

**Solution:** Use query complexity analysis.

```go
import "github.com/99designs/gqlgen/graphql/handler/extension"

server := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

// Limit query complexity
server.Use(extension.FixedComplexityLimit(1000))  // Max 1000 complexity points

// Define complexity per field
type ComplexityRoot struct {
    Query struct {
        Deployments func(childComplexity int) int  // = childComplexity * 10
    }
}
```

**Why:** Prevent denial-of-service attacks.

---

### 6. Multi-Tenancy: Security Hardening

#### A. Namespace Isolation is Not Enough

**Problem:** Namespaces are soft isolation. One tenant can:
- Exhaust cluster CPU/memory (no ResourceQuotas)
- Read secrets from other namespaces (no RBAC)
- Deploy malicious code (no PodSecurityStandards)

**Solution:** Defense in depth.

```yaml
# 1. ResourceQuota (CPU, memory limits)
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-backend-quota
  namespace: team-backend
spec:
  hard:
    requests.cpu: "100"
    requests.memory: 200Gi
    count/deployments.scai.io: "50"

---
# 2. NetworkPolicy (isolate traffic)
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-cross-namespace
  namespace: team-backend
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: team-backend  # Only same namespace
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: team-backend

---
# 3. PodSecurityStandard (restrict privileged pods)
apiVersion: v1
kind: Namespace
metadata:
  name: team-backend
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

**Why:** Single control missing = security breach.

#### B. RBAC: Principle of Least Privilege

**Problem:** Give users `cluster-admin` = full access (can delete all deployments).

**Solution:** Granular roles.

```yaml
# Role: Can only manage own namespace's deployments
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: scai-deployer
  namespace: team-backend
rules:
  - apiGroups: ["scai.io"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["scai.io"]
    resources: ["deployments/status"]
    verbs: ["get", "watch"]  # Read-only status
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "list", "watch"]

---
# RoleBinding: Bind to team
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: team-backend-deployers
  namespace: team-backend
subjects:
  - kind: Group
    name: team-backend
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: scai-deployer
  apiGroup: rbac.authorization.k8s.io
```

**Test:**
```bash
# User in team-backend can create deployments
kubectl --as=user@team-backend.com create -f deployment.yaml
# ✅ Success

# User in team-backend CANNOT create in team-frontend
kubectl --as=user@team-backend.com create -f deployment.yaml -n team-frontend
# ❌ Error: forbidden
```

---

### 7. Cost Optimization

#### A. Estimate Before Deploy

**Problem:** User deploys expensive resources ($1000/month) by mistake.

**Solution:** Pre-deployment cost check.

```go
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    deployment := &scaiapi.Deployment{}
    r.Get(ctx, req.NamespacedName, deployment)

    // Estimate cost before creating infrastructure
    estimatedCost, err := r.CostEstimator.Estimate(decision)
    if err != nil {
        return ctrl.Result{}, err
    }

    // Check budget
    if deployment.Spec.Budget != nil {
        if estimatedCost > deployment.Spec.Budget.MaxMonthlyCost {
            r.Recorder.Event(deployment, corev1.EventTypeWarning, "BudgetExceeded",
                fmt.Sprintf("Estimated cost $%.2f exceeds budget $%.2f",
                    estimatedCost, deployment.Spec.Budget.MaxMonthlyCost))

            if deployment.Spec.Budget.EnforcementMode == "hard" {
                return ctrl.Result{}, fmt.Errorf("budget exceeded")
            }
        }
    }

    deployment.Status.EstimatedMonthlyCost = estimatedCost
    r.Status().Update(ctx, deployment)

    // Proceed with provisioning...
}
```

**Cost Estimator Implementation:**
```go
type CostEstimator struct {
    pricing map[string]float64  // instance_type -> hourly_cost
}

func (c *CostEstimator) Estimate(decision Decision) (float64, error) {
    switch decision.Strategy {
    case "vm":
        instanceCost := c.pricing[decision.Config["ec2_instance_type"]] * 730  // hours/month
        volumeCost := decision.Config["volume_size"] * 0.10  // $0.10/GB/month
        return instanceCost + volumeCost, nil
    case "kubernetes":
        nodes := decision.Config["eks_desired_nodes"]
        nodeCost := c.pricing[decision.Config["eks_node_type"]] * 730
        clusterCost := 0.10 * 730  // EKS control plane
        return float64(nodes)*nodeCost + clusterCost, nil
    }
}
```

#### B. Track Actual Costs (AWS Cost Explorer API)

**Problem:** Estimated cost != actual cost.

**Solution:** Periodic cost sync.

```go
// CronJob: Update actual costs daily
func (r *CostSyncer) SyncCosts(ctx context.Context) error {
    deployments := &scaiapi.DeploymentList{}
    r.Client.List(ctx, deployments)

    for _, deployment := range deployments.Items {
        // Query AWS Cost Explorer
        actualCost, err := r.AWS.GetCostByTags(map[string]string{
            "scai-deployment-id": deployment.UID,
        })
        if err != nil {
            log.Error(err, "Failed to fetch cost")
            continue
        }

        deployment.Status.ActualMonthlyCost = actualCost
        r.Status().Update(ctx, &deployment)
    }
}
```

**Grafana Dashboard:**
```
Estimated vs Actual Cost:
- Estimated: $500
- Actual: $450
- Savings: $50 (10%)
```

---

### 8. Observability: Key Metrics

#### Operator Metrics (Prometheus)

```go
var (
    // Counter: Total deployments
    deploymentsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "scai_deployments_total",
            Help: "Total deployments",
        },
        []string{"strategy", "framework", "outcome"},
    )

    // Histogram: Deployment duration
    deploymentDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "scai_deployment_duration_seconds",
            Help: "Deployment duration",
            Buckets: []float64{30, 60, 120, 300, 600, 1200},  // 30s to 20min
        },
        []string{"strategy"},
    )

    // Gauge: Active deployments
    activeDeployments = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "scai_active_deployments",
            Help: "Number of active deployments",
        },
        []string{"namespace", "strategy"},
    )

    // Histogram: RAG retrieval latency
    ragRetrievalLatency = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "scai_rag_retrieval_seconds",
            Help: "RAG retrieval latency",
            Buckets: prometheus.DefBuckets,
        },
    )
)

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    start := time.Now()
    defer func() {
        deploymentDuration.WithLabelValues(deployment.Status.Strategy).Observe(time.Since(start).Seconds())
    }()

    // ... reconciliation logic

    deploymentsTotal.WithLabelValues(
        deployment.Status.Strategy,
        deployment.Spec.Framework,
        string(deployment.Status.Phase),
    ).Inc()
}
```

**Grafana Queries:**
```promql
# Success rate
sum(rate(scai_deployments_total{outcome="success"}[5m])) /
sum(rate(scai_deployments_total[5m]))

# P95 deployment duration
histogram_quantile(0.95, rate(scai_deployment_duration_seconds_bucket[5m]))

# Active deployments by strategy
sum by (strategy) (scai_active_deployments)
```

---

## Common Pitfalls to Avoid

### 1. **Operator Infinite Loops**

**Symptom:** Operator reconciles every second, high CPU usage.

**Cause:** Status update triggers watch, which triggers reconcile, which updates status...

**Fix:** Use `generation` field to detect spec changes.

```go
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    deployment := &scaiapi.Deployment{}
    r.Get(ctx, req.NamespacedName, deployment)

    // Skip if already reconciled this generation
    if deployment.Status.ObservedGeneration == deployment.Generation {
        return ctrl.Result{}, nil
    }

    // ... reconciliation logic

    deployment.Status.ObservedGeneration = deployment.Generation
    r.Status().Update(ctx, deployment)
}
```

### 2. **RAG Retrieval Too Slow**

**Symptom:** LLM decision takes >10 seconds.

**Cause:** pgvector linear scan, no index.

**Fix:** See section 3.B (Vector Index Tuning).

### 3. **Crossplane Resource Stuck**

**Symptom:** `kubectl get infrastructure` shows `Synced: False` forever.

**Debug:**
```bash
# Check Crossplane events
kubectl describe infrastructure my-infra

# Check provider logs
kubectl logs -n crossplane-system deployment/provider-aws

# Check composed resources
kubectl get managed
```

**Common causes:**
- AWS credentials invalid
- IAM permissions missing
- Resource quota exceeded

### 4. **GraphQL Subscription Memory Leak**

**Symptom:** API server memory grows over time.

**Cause:** WebSocket connections never closed.

**Fix:** Set connection timeout.

```go
server := handler.NewDefaultServer(schema)
server.AddTransport(&transport.Websocket{
    KeepAlivePingInterval: 10 * time.Second,
    Upgrader: websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
    },
})
```

---

## Testing Strategy

### Unit Tests (Go)

```go
func TestDeploymentReconciler_Reconcile(t *testing.T) {
    // Use envtest (fake Kubernetes API)
    testEnv := &envtest.Environment{
        CRDDirectoryPaths: []string{"config/crd/bases"},
    }
    cfg, _ := testEnv.Start()
    defer testEnv.Stop()

    k8sClient, _ := client.New(cfg, client.Options{Scheme: scheme})

    reconciler := &DeploymentReconciler{
        Client: k8sClient,
        Analyzer: &MockAnalyzer{},
        RAG: &MockRAG{},
    }

    // Test: Create deployment, verify infrastructure created
    deployment := &scaiapi.Deployment{
        ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
        Spec: scaiapi.DeploymentSpec{
            Repository: scaiapi.Repository{URL: "https://github.com/test/app"},
        },
    }
    k8sClient.Create(context.Background(), deployment)

    _, err := reconciler.Reconcile(context.Background(), ctrl.Request{
        NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
    })
    assert.NoError(t, err)

    // Verify infrastructure created
    infra := &scaiapi.Infrastructure{}
    k8sClient.Get(context.Background(), types.NamespacedName{
        Name: "test-infra",
        Namespace: "default",
    }, infra)
    assert.NotNil(t, infra)
}
```

### Integration Tests (Ginkgo)

```go
var _ = Describe("Deployment Controller", func() {
    It("Should create infrastructure when deployment is created", func() {
        deployment := &scaiapi.Deployment{
            ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
            Spec: scaiapi.DeploymentSpec{
                Repository: scaiapi.Repository{URL: "https://github.com/test/app"},
            },
        }
        Expect(k8sClient.Create(ctx, deployment)).To(Succeed())

        Eventually(func() bool {
            infra := &scaiapi.Infrastructure{}
            err := k8sClient.Get(ctx, types.NamespacedName{
                Name: "test-infra",
                Namespace: "default",
            }, infra)
            return err == nil
        }, timeout, interval).Should(BeTrue())
    })
})
```

### E2E Tests (Real Cluster)

```bash
#!/bin/bash
# e2e-test.sh

# 1. Create test deployment
kubectl apply -f - <<EOF
apiVersion: scai.io/v1alpha1
kind: Deployment
metadata:
  name: e2e-test-flask
  namespace: default
spec:
  repository:
    url: https://github.com/user/flask-app
  prompt: "Deploy Flask app"
  cloudProvider: aws
  region: us-east-1
EOF

# 2. Wait for deployment to be ready
kubectl wait --for=condition=Ready deployment/e2e-test-flask --timeout=600s

# 3. Verify infrastructure created
kubectl get infrastructure e2e-test-flask-infra -o yaml

# 4. Check application is accessible
APP_URL=$(kubectl get deployment e2e-test-flask -o jsonpath='{.status.endpoints[0].url}')
curl -f $APP_URL || exit 1

# 5. Cleanup
kubectl delete deployment e2e-test-flask
```

---

## Performance Benchmarks

### Target Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Deployment Time (VM) | < 5 minutes | Time from CR creation to status=Ready |
| Deployment Time (K8s) | < 10 minutes | Includes EKS cluster creation |
| RAG Retrieval Latency | < 100ms | Time to retrieve top-5 similar deployments |
| LLM Decision Time | < 5 seconds | Time for LLM to return strategy |
| GraphQL Query Latency (list) | < 200ms | Time to list 100 deployments |
| Operator CPU Usage (idle) | < 50m | CPU when no reconciliations |
| Operator Memory Usage | < 200Mi | Typical memory footprint |

### Benchmark Script

```bash
#!/bin/bash
# benchmark.sh

# 1. Deploy 10 Flask apps
for i in {1..10}; do
  kubectl apply -f - <<EOF
apiVersion: scai.io/v1alpha1
kind: Deployment
metadata:
  name: bench-flask-$i
spec:
  repository:
    url: https://github.com/user/flask-app
  prompt: "Deploy Flask app"
EOF
  START=$(date +%s)

  # Wait for ready
  kubectl wait --for=condition=Ready deployment/bench-flask-$i --timeout=600s

  END=$(date +%s)
  DURATION=$((END - START))
  echo "Deployment $i took ${DURATION}s"
done

# 2. Measure average
# Target: <300s (5 minutes)
```

---

## Conclusion

**Key Takeaways:**

1. ✅ **Start small**: Basic operator → prove concept → iterate
2. ✅ **Use frameworks**: Kubebuilder, gqlgen (don't reinvent)
3. ✅ **Idempotency**: Test by applying same CR multiple times
4. ✅ **Finalizers**: Ensure cleanup (no orphaned resources)
5. ✅ **RAG tuning**: Index, re-ranking, chunking
6. ✅ **Security**: Defense in depth (RBAC + NetworkPolicy + PodSecurity)
7. ✅ **Observability**: Prometheus metrics from day 1
8. ✅ **Testing**: Unit + Integration + E2E
9. ✅ **Benchmarks**: Measure performance continuously

**Document Version:** 1.0
**Last Updated:** 2025-10-18
