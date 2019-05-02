package data

//var PRODUCTS map[string]Product	//maps product id to product struct. Datastore for products
//var USERS map[string]bool //stores hash of RSA public key as key. Used by miners to check if user exists

type Product struct {
	ProductName string
	ProductID   string //GTIN of product, must be provided by merchant by checking the product
}

type Review struct {
	Product Product
	Review  string
}

type Users struct {
	UserSet map[string]bool //stores hash of RSA public key as key. Used by miners to check if user exists
}

type Products struct {
	ProductSet map[string]Product //maps product id to product struct. Datastore for products
}

type Transaction struct {
	TransactionID string //This is essentially a hash of the transaction object
	PublicKey     string
	ReviewObj     Review
}

type UserWebUpload struct {
	PublicKey string
	ProductID string
	Review    string
	Signature string
}

type SignMessage struct {
	PrivateKey string
	ProductID  string
	Review     string
}
