# MallBots

<hr>

###### ⚔️ References

# Postman Collections

- [gRPC](https://gold-shuttle-395606.postman.co/workspace/My-Workspace~e9564e49-df76-48b9-8f40-1c74ee320241/collection/687cdfafaff3c5780e7dbdd7?action=share&creator=10413281&active-environment=10413281-37d0952d-6a07-443b-a8f3-83805f295a77)
- [REST](https://gold-shuttle-395606.postman.co/workspace/My-Workspace~e9564e49-df76-48b9-8f40-1c74ee320241/collection/10413281-2748aa7f-effe-4b38-8fd4-a68d044657c0?action=share&source=copy-link&creator=10413281)

<hr>

- Swagger UI: [http://localhost:8080/](http://localhost:8080/)
- HTTP service: [http://localhost:8080/](http://localhost:8080/)
- gRPC service: [grpc://127.0.0.1:8085](grpc://127.0.0.1:8085)

# System Design

```mermaid
---
title: Mallbots System Design

config:
  flowchart:
    htmlLabels: false
---
flowchart RL
  customers[Customers Service]
  notifications[Notifications Service]
  payments[Payments Service]
  ordering[Ordering Service]
  depot[Depot Service]
  stores[Stores Service]
  baskets[Baskets Service]
  database[(Database)]

  customers_doc["
  RegisterCustomer
  AuthorizeCustomer
  GetCustomer
  EnableCustomer
  DisableCustomer
  "]@{ shape: doc }

  notifications_doc["
  NotifyOrderCreated
  NotifyOrderCanceled
  NotifyOrderReady
  "]@{ shape: doc }

  payments_doc["
  AuthorizePayment
  ConfirmPayment
  CreateInvoice
  AdjustInvoice
  PayInvoice
  CancelInvoice
  "]@{ shape: doc }

  ordering_doc["
  CreateOrder
  GetOrder
  CancelOrder
  ReadyOrder
  CompleteOrder
  "]@{ shape: doc }

  depot_doc["
  CreateShoppingList
  CancelShoppingList
  AssignShoppingList
  CompleteShoppingList
  "]@{ shape: doc }

  stores_doc["
  CreateStore
  GetStore
  GetStores
  EnableParticipation
  DisableParticipation
  GetParticipatingStores
  AddProduct
  RemoveProduct
  GetCatalog
  GetProduct
  "]@{ shape: doc }

  baskets_doc["
  StartBasket
  CancelBasket
  CheckoutBasket
  AddItem
  RemoveItem
  GetBasket
  "]@{ shape: doc }

  customers_doc -...- customers
  notifications_doc -...- notifications
  payments_doc -...- payments
  ordering_doc -...- ordering
  depot_doc -...- depot
  stores_doc -...- stores
  baskets_doc -...- baskets

  notifications -->|gRPC| customers
  customers --> database
  ordering -->|gRPC| notifications
  ordering -->|gRPC| customers
  ordering -->|gRPC| payments
  ordering -->|gRPC| depot
  ordering --> database
  payments -->|gRPC| ordering
  payments --> database
  depot -->|gRPC| ordering
  depot -->|gRPC| stores
  depot --> database
  stores --> database
  baskets -->|gRPC| ordering
  baskets -->|gRPC| stores
  baskets --> database
```
