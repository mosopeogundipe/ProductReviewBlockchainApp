# Blockchain Product Review System

### Link to the youtube video demo: https://youtu.be/xT9bUAuePCw

### Why Blockchain

People find it hard to trust reviews from most review websites. Using blockchain would allow us achieve trust through immutability and transparency that blockchain provides. Blockchain is also good because it allows reviewers to remain anonymous, while ensuring they are verified. This is useful for reviews of sensitive products. This solution will only work only for products that have a unique product id (number usually found on a barcode). This ID might be same for products in the same batch, and that’s still okay for this solution

### How

1. Trust - There will be a centralized database where merchants register their product and the quantity available. Users will need to submit their account id and product id/barcode of the product if it's to be valid. Also ensure they can only leave one review per product by checking against a centralized database of product reviews
Note: Allowing only people we verify on our central DB to post to the blockchain gives us trust. 
2. Immutability - This can be used to get people to trust review sites more, because no product owner will be able to influence or change a review after it’s posted
3. Transparency - users, merchants and third parties will be able to see the entire activities on the system

### Architecture

All interactions in this project are done via web APIs. I'm not using any html pages.

****ARCHITECTURE DIAGRAM WILL BE INCLUDED SOON**********

My architecture comprises:
1) Web Client - web browser or clients like PostMan for accessing the APIs this solution. User actions are also performed by calling the user related APIs from this web client 
2) Middle Layer - Is the middleman between the miners and users. Contains the APIs for communication between users and miners, and vice versa. Also contains data structures and acts as a centralized data store for data that is stored centrally, e.g. list of products.
3) Miners Layer - This is the layer where the miners (peers) in the network reside and communicate using gossip protocol

### Functionalities

#### 1. User Creation
This entails the user calling a GET /register/user API in the middle layer, from a web client. This API creates a RSA Public-Private key pair, and stores the hash of the public key in a Set, so that it's easy to validate if a user has been created in future. The API then returns this Public-Private key pair to the user (web client).

#### 2. Product Creation
This entails calling a POST /register/product API in the middle layer, from a web client. Although anyone can perform this action right now, this should ideally be done by a seller. This API accepts JSON input of the form:

<pre>
{
"ProductName": "Smart Water",
"ProductID": "A0372926671"
}
</pre>

It stores this JSON input in a Product Object structure, and maps the ProductID to Product Object for easy retrieval. See the data structures below:

<pre>
type Product struct {
	ProductName string
	ProductID   string 
}


//maps product id to product struct. Datastore for products
type Products struct {
	ProductSet map[string]Product 
}
</pre>

The API returns a message that says if registration was successful or not.

#### 3. Miner Registration
This entails calling a POST /register/miners API in the middle layer, from a web client, so that the miners can register their IP addresses on the middle layer. The middle layer would need to forward product reviews to miners and as such would need the miner IPs. The API returns a status code of 200 if successful. For this to work, a RegisterInMiddleLayer() function was added at miners (peers) end which the miners use to provide the IPs to this Miner Registration API. Each miner must call that function successfully during startup.

#### 4. Sign Message
This entails calling a POST /sign/message API in the middle layer, from a web client. It's used for signing the message that would be sent by users to the miners, and is the first step of message validation. The API accepts a JSON body like:

<pre>
{
"PrivateKey": "xhsxhsgxshkxgshdsygsjkxxvvssjkxjhsxjxvbxvcv",
"ProductID": "A0372926671",
"Review": "this was a horrible product"
}
</pre>

It verifies that the ProductID exists in the product database (number 2 above), and if so it uses these details to create a transaction object with blank public key and blank transaction id, signs the transaction object and returns the signature string to the web client. 

<pre>
//Transaction Object:
type Transaction struct {
	TransactionID string //This is essentially a hash of the transaction object
	PublicKey     string
	ReviewObj     Review
}

//Review Object
type Review struct {
	Product Product
	Review  string
}

//Product Object
type Product struct {
	ProductName string
	ProductID   string //GTIN of product, must be provided by merchant by checking the product
}
</pre>

#### 5. Post Review
This entails calling a POST /review/post API in the middle layer, from a web client. It's used for accepting a review from a user. The API accepts JSON input that looks like:

<pre>
{
"PublicKey": "xhsxhsgxshkxgshdsygsjkxxvvssjkxjhsxjxvbxvcv",
"ProductID": "A0372926671",
"Review": "this was a horrible product"
"Signature": "edhgjehj372wuowhs02wio290w2wshjhs761278769821sfjsfsjhg387268972"
}
</pre>

"Signature" above is the signature for this same message, which must have been created prior by performing the steps in no 4 above.
The API verifies that the user and product exists by using checking the user database (step 1 above) for the hash of the public key, and checking the product database (step 2 above) for the Product ID. If both exist, it creates a transaction object (refer to step 4 above) and calls a POST /transaction/receive API for every miner in list of miners from the middle layer. To this API, it sends the JSON string of the transaction object in the API request body and the Signature as query strings in the url.

#### 6. Receive Transaction
This POST /transaction/receive API resides at the miners' end and is called from step 5 above. Here, each miner verifies the message was signed by the user who sent this request. It does so using the Public Key, the Message signature and the Message itself (the message is the transaction object). 
If the message verification is successful, it adds the transaction object to the miner's pool.

#### 7. Ensuring users can submit only one review per product
This check as part of the logic in the POST /transaction/receive API (called in step 6 above). It hashes the user's public key and checks through the MPT of every block in the blockchain to see if it this hashed public key exists as a key in the MPT. If the public key exists in a block's MPT, it checks the productID of the transaction and confirms if it's the same as the productID of the review being submitted. If there's an existing record, it returns an error message explaining that only one review can be submitted per user

#### 8. Miners creating blocks based on submitted reviews in transaction pool
Changed the StartTryingNonces() function in the p3/handlers class to generate a new block every 5 seconds (this time is configurable). Every 5 seconds, if there are transactions in the transaction pool, it takes each transaction, adds them to mpt where hash of the user's public key is the KEY and transaction JSON is the VALUE. It then proceeds to generate nonce and create block using already existing functionalities from previous projects.

#### 9. Tracking of all reviews for all products based on user id
This entails calling a POST /reviews/find/publickey API from a web client. This is hosted on the miners' end as each miner has a separate copy of the blockchain. The API accepts JSON input that looks like:

<pre>
{
"PublicKey": "xhsxhsgxshkxgshdsygsjkxxvvssjkxjhsxjxvbxvcv",
"ProductID": "",
"Review": ""
"Signature": ""
}
</pre>

It uses the same input structure as previous APIs but expects that only the PublicKey is populated while other fields come with an empty string. It hashes the user's public key and checks through the MPT of every block in the blockchain to see if it this hashed public key exists as a key in the MPT. If locates the value of the key in every MPT and returns all the values to the user. Note that the value is a Transaction Object, so it returns all those details to the user who calls this API. If no record is found, it returns a message that says so.

#### 10. Tracking of all reviews for a product based on the product id and user id
This entails calling a POST /reviews/find/all API from a web client. This is hosted on the miners' end as each miner has a separate copy of the blockchain. The API accepts JSON input that looks like:

<pre>
{
"PublicKey": "xhsxhsgxshkxgshdsygsjkxxvvssjkxjhsxjxvbxvcv",
"ProductID": "A36282722898",
"Review": ""
"Signature": ""
}
</pre>

It uses the same input structure as previous APIs but expects that both the PublicKey and ProductID are populated while other fields come with an empty string. It hashes the user's public key and checks through the MPT of every block in the blockchain to see if it this hashed public key exists as a key in the MPT. If locates the value of the key in every MPT and returns all the values to the user if they also have a matching product ID with the one sent from the API request. Note that the value is a Transaction Object, so it returns all those details to the user who calls this API. If no record is found, it returns a message that says so.
