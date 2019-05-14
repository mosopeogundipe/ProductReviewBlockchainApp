package logic

import (
	"cs686-blockchain-p3-mosopeogundipe/p5/data"
)

func AddProduct(product data.Product, products *data.Products) (bool, string) {
	if len(products.ProductSet) == 0 {
		products.ProductSet = make(map[string]data.Product)
		if _, exists := products.ProductSet[product.ProductID]; !exists { //add only if ProductID doesn't already exist in map
			products.ProductSet[product.ProductID] = product
		} else {
			return false, "exists"
		}
	} else {
		if _, exists := products.ProductSet[product.ProductID]; !exists { //add only if ProductID doesn't already exist in map
			products.ProductSet[product.ProductID] = product
		} else {
			return false, "exists"
		}
	}
	return true, ""
}
