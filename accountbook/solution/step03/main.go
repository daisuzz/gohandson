// STEP03: データの記録

package main

import "fmt"

// 品目と値段を一緒に扱うために
// Itemという構造体の型を定義する
type Item struct {
	// Categoryは文字列型のフィールド
	Category string
	// Priceは整数型のフィールド
	Price int
}

func main() {

	// Item型のitemという名前の変数を定義する
	var item Item

	fmt.Print("品目>")
	// 入力した値をitemのCategoryフィールドに入れる
	fmt.Scan(&item.Category)

	fmt.Print("値段>")
	// 入力した値をitemのPriceフィールドに入れる
	fmt.Scan(&item.Price)

	// 品目に「コーヒー」、値段に「100」と入力した場合に
	// 「コーヒーに100円使いました」と表示する
	fmt.Printf("%sに%d円使いました\n", item.Category, item.Price)
}
