package llm

// DeploymentKnowledgeBase contains expert knowledge about deployment patterns
const DeploymentKnowledgeBase = `# Cloud Deployment Expert Knowledge Base

## Framework Characteristics

### Python Web Frameworks

**Flask**
- Typical Memory: 256MB - 512MB
- Startup Time: 5-10 seconds
- Concurrency: Low-Moderate (WSGI)
- Common Use: Web apps, APIs, microservices
- Default Port: 5000
- Production Server: Gunicorn, uWSGI
- Best Deployment: VM or Kubernetes

**Django**
- Typical Memory: 512MB - 1GB
- Startup Time: 10-15 seconds
- Concurrency: Low-Moderate (WSGI)
- Common Use: Full-stack web apps, admin panels
- Default Port: 8000
- Production Server: Gunicorn + Nginx
- Best Deployment: VM or Kubernetes
- Special Needs: Static files, database migrations

**FastAPI**
- Typical Memory: 128MB - 256MB
- Startup Time: 2-5 seconds
- Concurrency: High (ASGI)
- Common Use: APIs, microservices
- Default Port: 8000
- Production Server: Uvicorn, Gunicorn
- Best Deployment: VM, Kubernetes, or Serverless

### JavaScript/Node.js Frameworks

**Express**
- Typical Memory: 128MB - 256MB
- Startup Time: 1-3 seconds
- Concurrency: High (event loop)
- Common Use: APIs, web servers
- Default Port: 3000
- Production: PM2, Node cluster
- Best Deployment: VM, Kubernetes, or Serverless

**Next.js**
- Typical Memory: 256MB - 512MB
- Startup Time: 5-10 seconds
- Concurrency: High
- Common Use: React apps, SSR
- Default Port: 3000
- Production: Next.js server
- Best Deployment: VM or Kubernetes

### Other Languages

**Go**
- Typical Memory: 20MB - 100MB
- Startup Time: < 1 second
- Concurrency: Very High (goroutines)
- Common Use: APIs, microservices, tools
- Default Port: 8080
- Production: Direct binary
- Best Deployment: Kubernetes or Serverless

**Ruby (Rails)**
- Typical Memory: 512MB - 1GB
- Startup Time: 15-30 seconds
- Concurrency: Low (MRI), High (JRuby)
- Common Use: Full-stack web apps
- Default Port: 3000
- Production: Puma, Passenger
- Best Deployment: VM or Kubernetes

## Deployment Strategy Decision Rules

### Choose VM (EC2) when:
✓ Traditional web framework (Flask, Django, Rails)
✓ No containerization (no Dockerfile)
✓ Simple single-service application
✓ Needs persistent local file system
✓ Long-running background processes
✓ Dependencies < 15 packages
✓ Development/staging environments
✓ Quick deployment preferred

Instance Sizing:
- t3.micro (1 vCPU, 1GB): Flask, Express, simple apps
- t3.small (2 vCPU, 2GB): Django, Rails, moderate traffic
- t3.medium (2 vCPU, 4GB): Production apps, high traffic

### Choose Kubernetes (EKS) when:
✓ Has Dockerfile present
✓ Has docker-compose.yml (multi-service)
✓ Microservices architecture
✓ Need horizontal auto-scaling
✓ Need high availability (multi-AZ)
✓ Complex networking requirements
✓ Dependencies > 20 packages
✓ Production workload
✓ Need rolling updates/canary deployments
✓ Multiple environments (dev/stage/prod)

### Choose Serverless (Lambda) when:
✓ Stateless API (REST/GraphQL)
✓ No local file system writes needed
✓ Execution time < 15 minutes
✓ Sporadic/unpredictable traffic
✓ Cost optimization priority
✓ Minimal dependencies (< 5 packages)
✓ Frameworks: FastAPI, Express (simple)
✓ Cold start acceptable (2-5 seconds)

### Anti-Patterns (What NOT to do):

❌ Don't use Lambda for:
- Long-running processes (> 15 min)
- WebSocket servers
- File uploads/processing (> 6MB)
- Frameworks needing local state (Django sessions)

❌ Don't use Kubernetes for:
- Single simple service
- No containerization
- Small team (< 3 developers)
- MVP/prototype stage

❌ Don't use VM for:
- Microservices (> 5 services)
- Need for auto-scaling
- Container-native apps

## AWS Configuration Best Practices

### Security Groups
- Restrict SSH (22) to known IPs, not 0.0.0.0/0
- Open only application port (80, 443, 3000, 5000, 8000)
- Allow all outbound (for package installs)

### Application Configuration
- Bind to 0.0.0.0 (not localhost/127.0.0.1)
- Use environment variables for config
- Never hardcode secrets
- Set proper CORS headers

### Performance
- Use production servers (Gunicorn, uWSGI, PM2)
- Enable gzip compression
- Set worker processes = CPU cores
- Configure health check endpoints

### Monitoring
- CloudWatch logs for application logs
- CloudWatch metrics for CPU/memory
- Application-level metrics (requests/sec)
- Set up alarms for high CPU/memory

## Common Port Mappings
- Flask: 5000
- Django: 8000
- FastAPI: 8000
- Express: 3000
- Next.js: 3000
- Go: 8080
- Rails: 3000
- Streamlit: 8501

## Dependency Analysis

### Low Dependencies (< 5):
Usually simple apps, good for Serverless or VM

### Medium Dependencies (5-20):
Typical web apps, good for VM

### High Dependencies (> 20):
Complex apps, consider Kubernetes for better management

## File System Requirements

### Ephemeral Storage OK:
- APIs with no file uploads
- Stateless applications
- Temporary caching
→ Can use Serverless

### Persistent Storage Needed:
- File uploads
- Local SQLite database
- Generated reports
- Session storage
→ Use VM or Kubernetes with PVC
`

// FewShotExamples contains example deployment decisions
const FewShotExamples = `# Example Deployment Decisions

## Example 1: Simple Flask Hello World
**Application:**
- Framework: Flask
- Language: Python
- Dependencies: 3 (Flask, requests, gunicorn)
- Has Dockerfile: No
- Has docker-compose: No
- Port: 5000
- Stateful: No

**Decision: VM (t3.micro)**
**Reasoning:** Simple web application with minimal dependencies. No containerization present. Traditional deployment suitable. Quick to deploy and maintain.

---

## Example 2: Express Microservices Platform
**Application:**
- Framework: Express
- Language: JavaScript
- Dependencies: 25 (express, redis, pg, bull, winston, etc.)
- Has Dockerfile: Yes
- Has docker-compose: Yes (4 services: app, redis, postgres, nginx)
- Port: 3000
- Stateful: Yes (database, cache)

**Decision: Kubernetes (EKS)**
**Reasoning:** Multi-service architecture with docker-compose present. Already containerized. High dependency count indicates complex application. Needs orchestration for service discovery and scaling.

---

## Example 3: FastAPI Simple REST API
**Application:**
- Framework: FastAPI
- Language: Python
- Dependencies: 2 (fastapi, pydantic)
- Has Dockerfile: No
- Has docker-compose: No
- Port: 8000
- Stateful: No (pure API, no database)

**Decision: Serverless (Lambda + API Gateway)**
**Reasoning:** Stateless REST API with minimal dependencies. No file system requirements. Cost-effective for variable traffic. Fast startup time with FastAPI.

---

## Example 4: Django E-commerce Site
**Application:**
- Framework: Django
- Language: Python
- Dependencies: 15 (Django, psycopg2, pillow, celery, redis, etc.)
- Has Dockerfile: Yes
- Has docker-compose: No
- Port: 8000
- Stateful: Yes (database, media files)

**Decision: VM (t3.small)**
**Reasoning:** Traditional Django app with moderate complexity. Has Dockerfile but no orchestration needs. Requires persistent storage for media files. VM provides simple deployment with necessary persistence.

---

## Example 5: Next.js React Application
**Application:**
- Framework: Next.js
- Language: JavaScript
- Dependencies: 30+ (next, react, many UI libraries)
- Has Dockerfile: Yes
- Has docker-compose: No
- Port: 3000
- Stateful: No (SSR, API routes)

**Decision: VM (t3.small) or Kubernetes**
**Reasoning:** Containerized SSR application. If single instance needed → VM. If need scaling for traffic spikes → Kubernetes. High dependency count manageable in both.

---

## Example 6: Go Microservice
**Application:**
- Framework: Go (net/http)
- Language: Go
- Dependencies: 5 (minimal standard library usage)
- Has Dockerfile: Yes
- Has docker-compose: No
- Port: 8080
- Stateful: No

**Decision: Kubernetes or Serverless**
**Reasoning:** Lightweight containerized microservice. Fast startup makes it Lambda-friendly. If part of larger system → Kubernetes. If standalone API → Lambda.

---

## Example 7: Python Data Processing Script
**Application:**
- Framework: None (script)
- Language: Python
- Dependencies: 10 (pandas, numpy, boto3, etc.)
- Has Dockerfile: No
- Scheduled: Yes (cron)
- Execution: 5-30 minutes

**Decision: VM with cron or Lambda (scheduled)**
**Reasoning:** Batch processing workload. If < 15 min → Lambda with EventBridge. If > 15 min → VM with cron. VM provides more flexibility for long-running jobs.
`

// DecisionPromptTemplate is the template for the final decision prompt
const DecisionPromptTemplate = `Based on the knowledge base and examples above, analyze this new application:

**User Request:** %s

**Application Analysis:**
- Framework: %s
- Language: %s
- Dependencies: %d packages
- Has Dockerfile: %v
- Has docker-compose: %v
- Port: %d
- Start Command: %s
- Estimated Memory: %s

**Your Task:**
Recommend the BEST deployment strategy for this application.

**Response Format:**
STRATEGY: <vm|kubernetes|serverless>
REASON: <one sentence explanation>

Respond now:
`
