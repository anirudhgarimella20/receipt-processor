# receipt-processor
Docker commands to run the service:

 $ docker build --tag service . 
 
 $ docker run --publish 8080:8080 service

 POST Method:

curl --location 'localhost:8080/receipts/process' \
--header 'Content-Type: application/json' \
--data '{
  "retailer": "M&M Corner Market",
  "purchaseDate": "2022-03-20",
  "purchaseTime": "14:33",
  "items": [
    {
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    },{
      "shortDescription": "Gatorade",
      "price": "2.25"
    }
  ],
  "total": "9.00"
}'


GET Method : 

curl --location 'localhost:8080/receipts/{id}/points'
 
