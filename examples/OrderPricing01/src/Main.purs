module Main where

import Prelude

import Effect (Effect)
import Effect.Console (log)
import OrderPricing.Input (loadOrder)
import OrderPricing.Pricing (priceOrder)
import OrderPricing.Render (renderPricingResult)

main :: Effect Unit
main = log (renderPricingResult (priceOrder (loadOrder 1)))
