package postgres

const (
	getInventory    = "SELECT * FROM inventory order by art_id"
	insertProduct   = "INSERT INTO product (product_name, art_id, amount) VALUES ($1,$2,$3)"
	insertStock     = "INSERT INTO inventory(art_id, art_name, stock) VALUES ($1,$2,$3)"
	getProductStock = "SELECT pr.product_name, min(i.stock/pr.amount) as available_product FROM product pr,inventory i WHERE pr.art_id=i.art_id GROUP BY pr.product_name ORDER BY pr.product_name"
	updateSaleInfo  = "UPDATE inventory i SET stock=stock-1 from product pr WHERE pr.art_id= i.art_id and stock>= 1 AND pr.product_name=$1"
	inStock         = "SELECT count(*) from product pr, inventory i WHERE pr.art_id=i.art_id AND pr.product_name = $1 AND i.stock=0"
	productExist    = "select count(*) from product where product_name=$1"
)
