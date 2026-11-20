module Main where

import Prelude

import Data.Maybe (Maybe(..))
import Effect (Effect)
import Effect.Console (log)
import Run (Run)
import Run as Run
import Type.Proxy (Proxy(..))
import Type.Row (type (+))

data RenderF a = Fill Number a

derive instance Functor RenderF

data MetricsF a = Measure String (Number -> a)

derive instance Functor MetricsF

type RENDER r = (render :: RenderF | r)
type METRICS r = (metrics :: MetricsF | r)

_render = Proxy :: Proxy "render"
_metrics = Proxy :: Proxy "metrics"

newtype Draw a = Draw (Run (METRICS + RENDER + ()) a)

unDraw :: forall a. Draw a -> Run (METRICS + RENDER + ()) a
unDraw (Draw r) = r

derive newtype instance Functor Draw
derive newtype instance Apply Draw
derive newtype instance Applicative Draw
derive newtype instance Bind Draw
derive newtype instance Monad Draw

fill :: Number -> Draw Unit
fill n = Draw (Run.lift _render (Fill n unit))

shape :: Maybe Number -> Draw Unit
shape = case _ of
  Just n -> fill n
  Nothing -> pure unit

main :: Effect Unit
main = do
  let program = unDraw (shape (Just 2.0) *> shape Nothing)
  Run.interpret (Run.case_ # Run.on _render (\(Fill n next) -> log (show n) *> pure next) # Run.on _metrics (\(Measure _ k) -> pure (k 1.0))) program
  log "done"
