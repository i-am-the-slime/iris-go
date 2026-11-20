module OrderPricing.Pricing (priceOrder) where

import Prelude

import Data.Array as Array
import Data.Foldable (foldl)
import Data.Maybe (Maybe(..))
import OrderPricing.Domain
  ( CustomerTier(..)
  , Order
  , OrderLine
  , PricingError(..)
  , PricingResult(..)
  )

type PricingAccumulator =
  { subtotal :: Int
  , itemCount :: Int
  , firstError :: Maybe PricingError
  }

accumulateLine :: PricingAccumulator -> OrderLine -> PricingAccumulator
accumulateLine accumulator line = case accumulator.firstError of
  Just _ -> accumulator
  Nothing
    | line.quantity <= 0 ->
        accumulator { firstError = Just (InvalidQuantity line.sku line.quantity) }
    | line.unitPrice < 0 ->
        accumulator { firstError = Just (InvalidUnitPrice line.sku line.unitPrice) }
    | otherwise ->
        accumulator
          { subtotal = accumulator.subtotal + line.quantity * line.unitPrice
          , itemCount = accumulator.itemCount + line.quantity
          }

discountBasisPoints :: CustomerTier -> Int
discountBasisPoints Regular = 0
discountBasisPoints Preferred = 500
discountBasisPoints Wholesale = 1500

priceOrder :: Order -> PricingResult
priceOrder order
  | Array.null order.lines = PricingFailed EmptyOrder
  | otherwise =
      let
        accumulator = foldl accumulateLine
          { subtotal: 0, itemCount: 0, firstError: Nothing }
          order.lines
      in
        case accumulator.firstError of
          Just pricingError -> PricingFailed pricingError
          Nothing ->
            let
              discount = accumulator.subtotal * discountBasisPoints order.customerTier / 10000
            in
              Priced
                { orderId: order.orderId
                , itemCount: accumulator.itemCount
                , subtotal: accumulator.subtotal
                , discount
                , total: accumulator.subtotal - discount
                }
