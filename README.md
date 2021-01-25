# Warehouse software

Warehouse is an application that the inventory related processes can be handled. There are four main functionalities, explained in functionalities section,are provided so far. Github actions are used as CI/CD pipeline and the service is deployed to Google Cloud Run. To maintain data storage needs of service Google Cloud SQL is used.

##Preparation 
### Github action
There are two actions called Go and Deploy To Cloud Run. Go action is responsible to check build and tests. Other action is the one that deploys the service to cloud.
For the deployment pipeline there are three secrets needs to be added in Secret from repository settings as `DBHOST`, `GCP_PROJECT_ID`, `GCP_SA_KEY_JSON`.
All action yml files can be found under workflows folder.

![Go](https://github.com/auknl/warehouse/workflows/Go/badge.svg?branch=main)
![Deploy to Cloud Run](https://github.com/auknl/warehouse/workflows/Deploy%20to%20Cloud%20Run/badge.svg?branch=main)

### Google Cloud Run
To be able use Google Cloud systems first an account and a new project needed. Google provides up to $300 free credits to explore its services.   
For the configurations of project and service there are two reference pages are used.
* https://dev.to/pcraig3/quickstart-continuous-deployment-to-google-cloud-run-using-github-actions-fna
* https://medium.com/google-cloud/cloud-run-cloud-sql-6c8879ef96da

Addition to those, there are three critical points needs to be noted
* Private IP is created for GCSQL, Google SQL Proxy is the best practice but not used in this project.
* Serverless VPC have to be created with default settings. GCRun and GCSQL has to be in same VPC.
* For the SQL, Database creation can be done from console with just button clicking and table creation is done via Cloud Shell. 

### Endpoints
There are four main functionalities can be executed against the endpoint.

- Health check 
```
GET /warehouse/v1/health

```
------

- Get all Stock info from inventory.
```
GET /warehouse/v1/inventory

```
------
- Get all product stock that are available.
```
GET warehouse/v1/product

```
------

- Upload stock information of articles/items

```
POST warehouse/v1/inventory
RequestBody example: 

{
  "inventory": [
    {
      "art_id": "1",
      "name": "leg",
      "stock": "12"
    },
    {
      "art_id": "2",
      "name": "screw",
      "stock": "17"
    },
    {
      "art_id": "3",
      "name": "seat",
      "stock": "2"
    },
    {
      "art_id": "4",
      "name": "table top",
      "stock": "1"
    }
  ]
}

```
------
- Upload production information that maps production and its required items

```
POST warehouse/v1/product
RequestBody example: 

{
  "products": [
    {
      "name": "Dinning Table",
      "contain_articles": [
        {
          "art_id": "1",
          "amount_of": "4"
        },
        {
          "art_id": "2",
          "amount_of": "8"
        },
        {
          "art_id": "4",
          "amount_of": "1"
        }
      ]
    }
  ]
}

```
-----

- Sells the given product if it is in stock, and updates the stock info 

```
POST warehouse/v1/product/<Product Name>

```
-----

### How To Test
The endpoint url for the service is 
* https://warehouse-3klf3eut5a-ez.a.run.app

Curl commands for endpoints
* Health check
```
curl --request GET \
    --url https://warehouse-3klf3eut5a-ez.a.run.app/warehouse/v1/health
```
----

* GET inventory
```
curl --request GET \
   --url https://warehouse-3klf3eut5a-ez.a.run.app/warehouse/v1/inventory \
   --header 'Content-Type: application/json'
```
-----
   
* GET product stock
```
curl --request GET \
  --url https://warehouse-3klf3eut5a-ez.a.run.app/warehouse/v1/product \
  --header 'Content-Type: application/json'
```
-----

* POST upload inventory
```
curl --request POST \
  --url https://warehouse-3klf3eut5a-ez.a.run.app/warehouse/v1/inventory \
  --header 'Content-Type: application/json' \
  --data '{
  "inventory": [
    {
      "art_id": "1",
      "name": "leg",
      "stock": "12"
    },
    {
      "art_id": "2",
      "name": "screw",
      "stock": "17"
    },
    {
      "art_id": "3",
      "name": "seat",
      "stock": "2"
    },
    {
      "art_id": "4",
      "name": "table top",
      "stock": "1"
    }
  ]
}
'
```
-----

* POST upload product
```
curl --request POST \
  --url https://warehouse-3klf3eut5a-ez.a.run.app/warehouse/v1/product \
  --header 'Content-Type: application/json' \
  --data '{
  "products": [
    {
      "name": "Dining Chair",
      "contain_articles": [
        {
          "art_id": "1",
          "amount_of": "4"
        },
        {
          "art_id": "2",
          "amount_of": "8"
        },
        {
          "art_id": "3",
          "amount_of": "1"
        }
      ]
    },
    {
      "name": "Dinning Table",
      "contain_articles": [
        {
          "art_id": "1",
          "amount_of": "4"
        },
        {
          "art_id": "2",
          "amount_of": "8"
        },
        {
          "art_id": "4",
          "amount_of": "1"
        }
      ]
    }
  ]
}
'
```
----

* POST sell product -> Product Name is Dining Chair in the example
```
curl --request POST \
  --url https://warehouse-3klf3eut5a-ez.a.run.app/warehouse/v1/product/Dining%20Chair \
  --header 'Content-Type: application/json'
```

Here is my blog post about the project https://aysenur-ulubas90.medium.com/github-action-integration-with-gcr-for-go-application-be15176d837 
