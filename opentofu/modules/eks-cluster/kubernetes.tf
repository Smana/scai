# Kubernetes resources for application deployment

# Kubernetes Deployment
resource "kubernetes_deployment" "app" {
  depends_on = [module.eks]

  metadata {
    name = "${var.app_name}-deployment"
    labels = {
      app         = var.app_name
      environment = var.environment
      managed-by  = "scia"
    }
  }

  spec {
    replicas = var.replicas

    selector {
      match_labels = {
        app = var.app_name
      }
    }

    template {
      metadata {
        labels = {
          app         = var.app_name
          environment = var.environment
        }
      }

      spec {
        container {
          name  = var.app_name
          image = var.container_image

          port {
            container_port = var.application_port
            protocol       = "TCP"
          }

          env {
            name  = "APP_NAME"
            value = var.app_name
          }

          env {
            name  = "REGION"
            value = var.region
          }

          env {
            name  = "ENVIRONMENT"
            value = var.environment
          }

          # Resource requests and limits
          resources {
            requests = {
              cpu    = "100m"
              memory = "128Mi"
            }
            limits = {
              cpu    = "500m"
              memory = "512Mi"
            }
          }

          # Liveness probe to restart unhealthy containers
          liveness_probe {
            http_get {
              path = "/"
              port = var.application_port
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }

          # Readiness probe to control traffic routing
          readiness_probe {
            http_get {
              path = "/"
              port = var.application_port
            }
            initial_delay_seconds = 10
            period_seconds        = 5
            timeout_seconds       = 3
            failure_threshold     = 3
          }
        }

        # Security context for pod
        security_context {
          run_as_non_root = true
          run_as_user     = 1000
          fs_group        = 1000
        }
      }
    }
  }
}

# Kubernetes Service (LoadBalancer)
resource "kubernetes_service" "app" {
  depends_on = [kubernetes_deployment.app]

  metadata {
    name = "${var.app_name}-service"
    labels = {
      app         = var.app_name
      environment = var.environment
      managed-by  = "scia"
    }
    annotations = {
      "service.beta.kubernetes.io/aws-load-balancer-type" = "nlb"
    }
  }

  spec {
    type = "LoadBalancer"

    selector = {
      app = var.app_name
    }

    port {
      name        = "http"
      port        = 80
      target_port = var.application_port
      protocol    = "TCP"
    }
  }
}
