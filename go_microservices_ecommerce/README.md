# Go Ecommerce Store: Microservices E-Commerce Application

A modern, simple e-commerce application built with a microservice architecture. It demonstrates how to build and orchestrate lightweight services using Go, Vue.js, MariaDB (MySQL), and Kubernetes.

## 🚀 Architecture & Tech Stack

- **Frontend**: Vue.js 3, TypeScript, Vite, and TailwindCSS v4. Provides a premium, responsive UI.
- **Product Service**: Golang REST API (Port: `8081`). Manages the product catalog.
- **Order Service**: Golang REST API (Port: `8082`). Handles customer orders.
- **Database**: MySQL 8.0 (Internal Port: `3306`, Service Port: `8706`). Contains isolated logical databases for `products` and `orders`.
- **Orchestration**: Kubernetes (`k8s/` manifests included) designed for environments like Minikube.

## 📂 Project Structure

```text
go_microservices_ecommerce/
├── frontend/             # Vue 3 UI (Vite + Tailwind v4)
├── product-service/      # Go service for product management
├── order-service/        # Go service for order processing
├── k8s/                  # Kubernetes deployment manifests
└── .gitignore
```

## 🛠 Prerequisites

Before deploying the application, ensure you have the following installed:
- [Docker](https://docs.docker.com/get-docker/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- Node.js & npm (for local frontend development)
- Go 1.22+ (for local backend development)

## 📦 Deployment Instructions (Kubernetes/Minikube)

Follow these steps to deploy the application inside your local Minikube cluster:

### 1. Start Minikube & Label the Node
The frontend deployment requires a dedicated node label to schedule successfully.
```bash
minikube start
kubectl label nodes minikube dedicated=ecommerce-node --overwrite
```

### 2. Build Docker Images
Configure your shell to use Minikube's Docker daemon so the images are built directly into the cluster:
```bash
eval $(minikube docker-env)

# Build the images
docker build -t product-service:latest ./product-service
docker build -t order-service:latest ./order-service
docker build -t frontend:latest ./frontend
```

### 3. Apply Kubernetes Manifests
Apply all the configuration files to deploy the database, services, and frontend.
```bash
kubectl apply -f k8s/
```

### 4. Wait for Pods to Initialize
Ensure that all pods, especially the `ecommerce-db`, are running successfully:
```bash
kubectl get pods -w
```
*(Note: The Go microservices feature resilient retry logic and will automatically wait for the database to become ready).*

### 5. Expose Services via Port-Forwarding
Because the frontend runs in your browser but communicates with cluster-internal services, you must forward the ports:

Open separate terminal windows and run:
```bash
kubectl port-forward svc/frontend 8700:8700
kubectl port-forward svc/product-service 8081:8081
kubectl port-forward svc/order-service 8082:8082
```

### 6. View the Application
Navigate to [http://localhost:8700](http://localhost:8700) in your browser!

## 🔧 Local Development

If you prefer running the stack natively without Kubernetes:
1. Start a local MariaDB/MySQL server on port `8706`.
2. Run `go run main.go` in both `product-service` and `order-service` directories.
3. Run `npm install` and `npm run dev` in the `frontend` directory. 
