module OrderPricing.Render (renderPricingResult) where

import Prelude

import OrderPricing.Domain (PricingError(..), PricingResult(..))

renderPricingError :: PricingError -> String
renderPricingError EmptyOrder = "pricing-error=empty-order"
renderPricingError (InvalidQuantity sku quantity) =
  "pricing-error=invalid-quantity sku=" <> sku <> " quantity=" <> show quantity
renderPricingError (InvalidUnitPrice sku unitPrice) =
  "pricing-error=invalid-unit-price sku=" <> sku <> " unit-price=" <> show unitPrice

renderPricingResult :: PricingResult -> String
renderPricingResult (PricingFailed pricingError) = renderPricingError pricingError
renderPricingResult (Priced receipt) =
  "order=" <> receipt.orderId
    <> " items="
    <> show receipt.itemCount
    <> " subtotal="
    <> show receipt.subtotal
    <> " discount="
    <> show receipt.discount
    <> " total="
    <> show receipt.total
