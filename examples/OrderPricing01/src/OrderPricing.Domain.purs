module OrderPricing.Domain
  ( CustomerTier(..)
  , OrderLine
  , Order
  , Receipt
  , PricingError(..)
  , PricingResult(..)
  ) where

data CustomerTier
  = Regular
  | Preferred
  | Wholesale

type OrderLine =
  { sku :: String
  , quantity :: Int
  , unitPrice :: Int
  }

type Order =
  { orderId :: String
  , customerTier :: CustomerTier
  , lines :: Array OrderLine
  }

type Receipt =
  { orderId :: String
  , itemCount :: Int
  , subtotal :: Int
  , discount :: Int
  , total :: Int
  }

data PricingError
  = EmptyOrder
  | InvalidQuantity String Int
  | InvalidUnitPrice String Int

data PricingResult
  = PricingFailed PricingError
  | Priced Receipt
