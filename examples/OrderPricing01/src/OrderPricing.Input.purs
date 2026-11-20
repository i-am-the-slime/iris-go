module OrderPricing.Input (loadOrder) where

import OrderPricing.Domain (CustomerTier(..), Order, OrderLine)

foreign import orderNumber :: Int
foreign import preferredCustomer :: Boolean
foreign import bookSkuCode :: Int
foreign import bookQuantity :: Int
foreign import bookUnitPrice :: Int
foreign import penSkuCode :: Int
foreign import penQuantity :: Int
foreign import penUnitPrice :: Int

tier :: Int -> CustomerTier
tier 1 = Preferred
tier 2 = Wholesale
tier _ = Regular

orderIdFrom :: Int -> String
orderIdFrom 1007 = "ORD-1007"
orderIdFrom _ = "ORD-UNKNOWN"

skuFrom :: Int -> String
skuFrom 1 = "BOOK"
skuFrom 2 = "PEN"
skuFrom _ = "UNKNOWN"

loadOrder :: Int -> Order
loadOrder tierCode =
  let
    bookLine :: OrderLine
    bookLine =
      { sku: skuFrom bookSkuCode
      , quantity: bookQuantity
      , unitPrice: bookUnitPrice
      }

    penLine :: OrderLine
    penLine =
      { sku: skuFrom penSkuCode
      , quantity: penQuantity
      , unitPrice: penUnitPrice
      }
  in
    { orderId: orderIdFrom orderNumber
    , customerTier: tier (if preferredCustomer then tierCode else 0)
    , lines: [ bookLine, penLine ]
    }
