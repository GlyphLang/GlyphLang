# E-commerce Shopping Cart API

A complete shopping cart and product catalog system demonstrating complex business logic and calculations.

## Features

- Product catalog with categories
- Shopping cart management
- Add/update/remove items
- Automatic price calculations
- Tax calculation (8%)
- Shipping calculation (free over $100)
- Discount code support
- Stock management
- Order checkout process
- Rate limiting

## API Endpoints

### Products

#### Get All Products
```
GET /api/products?category=Electronics
```
Returns all products, optionally filtered by category.

**Response:**
```json
{
  "products": [
    {
      "id": 1,
      "name": "Laptop",
      "description": "High-performance laptop",
      "price": 999.99,
      "category": "Electronics",
      "stock": 10,
      "image_url": "https://...",
      "rating": 4.5
    }
  ],
  "total": 1,
  "category": "Electronics"
}
```

#### Get Product by ID
```
GET /api/products/:id
```
Returns details of a specific product.

#### Get Categories
```
GET /api/categories
```
Returns all product categories with item counts.

### Shopping Cart

#### Get Cart
```
GET /api/cart
```
Returns the current user's shopping cart. Creates an empty cart if none exists.

**Response:**
```json
{
  "id": 1,
  "user_id": 123,
  "items": [
    {
      "product_id": 1,
      "product_name": "Laptop",
      "quantity": 2,
      "unit_price": 999.99,
      "subtotal": 1999.98
    }
  ],
  "subtotal": 1999.98,
  "tax": 159.998,
  "shipping": 0.0,
  "discount": 0.0,
  "total": 2159.978,
  "updated_at": 1234567890
}
```

#### Add Item to Cart
```
POST /api/cart/items
```
Adds a product to the cart or updates quantity if already present.

**Request Body:**
```json
{
  "product_id": 1,
  "quantity": 2
}
```

**Response:**
```json
{
  "success": true,
  "message": "Item added to cart",
  "cart": {...}
}
```

**Validations:**
- Product must exist
- Quantity must be greater than 0
- Sufficient stock must be available

#### Update Cart Item Quantity
```
PUT /api/cart/items/:product_id
```
Updates the quantity of an item in the cart.

**Request Body:**
```json
{
  "quantity": 3
}
```

**Note:** Setting quantity to 0 will remove the item (alternatively use DELETE).

#### Remove Item from Cart
```
DELETE /api/cart/items/:product_id
```
Removes a specific item from the cart.

#### Clear Cart
```
DELETE /api/cart/clear
```
Removes all items from the cart.

#### Apply Discount Code
```
POST /api/cart/discount
```
Applies a discount code to the cart.

**Request Body:**
```json
{
  "code": "SUMMER2025"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Discount applied successfully",
  "cart": {...}
}
```

**Discount Types:**
- Percentage discounts (e.g., 20% off)
- Fixed amount discounts (e.g., $10 off)

#### Checkout
```
POST /api/cart/checkout
```
Creates an order from the cart and clears the cart.

**Response:**
```json
{
  "success": true,
  "message": "Order placed successfully",
  "data": {
    "id": 1,
    "cart_id": 1,
    "user_id": 123,
    "status": "pending",
    "total": 2159.978,
    "created_at": 1234567890
  }
}
```

## Type Definitions

### Product
```
: Product {
  id: int!
  name: str!
  description: str
  price: float!
  category: str!
  stock: int!
  image_url: str
  rating: float
}
```

### CartItem
```
: CartItem {
  product_id: int!
  product_name: str!
  quantity: int!
  unit_price: float!
  subtotal: float!
}
```

### Cart
```
: Cart {
  id: int!
  user_id: int!
  items: List[CartItem]
  subtotal: float!
  tax: float!
  shipping: float!
  discount: float!
  total: float!
  updated_at: timestamp
}
```

### Order
```
: Order {
  id: int!
  cart_id: int!
  user_id: int!
  status: str!
  total: float!
  created_at: timestamp
}
```

## Pricing Calculations

### Subtotal
Sum of all cart item subtotals (quantity × unit_price)

### Tax
8% of subtotal
```
tax = subtotal × 0.08
```

### Shipping
- $10.00 flat rate
- FREE for orders over $100

### Discount
Applied after subtotal calculation, before total

### Total
```
total = subtotal + tax + shipping - discount
```

## Business Rules

1. **Stock Management**
   - Cannot add more items than available stock
   - Stock is checked when adding/updating quantities

2. **Minimum Quantities**
   - Quantity must be greater than 0
   - Setting quantity to 0 removes the item

3. **Free Shipping**
   - Automatically applied when subtotal ≥ $100
   - Recalculated when cart is modified

4. **Discount Codes**
   - Must be active to be applied
   - Cannot exceed subtotal amount
   - Only one discount per cart

5. **Checkout Requirements**
   - Cart must contain at least one item
   - User must be authenticated
   - Stock is reserved/deducted on successful checkout

## Workflow Example

1. **Browse products:**
   ```
   GET /api/products?category=Electronics
   ```

2. **Add items to cart:**
   ```
   POST /api/cart/items
   {
     "product_id": 1,
     "quantity": 2
   }
   ```

3. **View cart:**
   ```
   GET /api/cart
   ```

4. **Apply discount:**
   ```
   POST /api/cart/discount
   {
     "code": "SAVE20"
   }
   ```

5. **Update quantity:**
   ```
   PUT /api/cart/items/1
   {
     "quantity": 3
   }
   ```

6. **Checkout:**
   ```
   POST /api/cart/checkout
   ```

## Error Responses

```json
{
  "success": false,
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

Common error codes:
- `NOT_FOUND` - Product or cart not found
- `VALIDATION_ERROR` - Invalid input data
- `OUT_OF_STOCK` - Insufficient stock
- `INVALID_CODE` - Invalid discount code
- `EXPIRED_CODE` - Discount code has expired
- `EMPTY_CART` - Cannot checkout empty cart

## Rate Limiting

- GET endpoints: 100-200 requests/minute
- POST endpoints: 10-50 requests/minute
- PUT endpoints: 50 requests/minute
- DELETE endpoints: 20-50 requests/minute

## Authentication

All cart operations require authentication:
- JWT token must be provided
- User ID is extracted from auth token
- Each user has their own cart
