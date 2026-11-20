module Snapshot.RecordCrossSyn01 (test) where

import Prelude

import CrossSyn.Types (Camera)

data List a = Nil | Cons a (List a)

-- `liftCamera` takes a `Camera` param — a record synonym from CrossSyn.Types,
-- NOT this module. It must survive as a real top-level function (forced by the
-- recursive `sumCams` calling it per element, so the optimizer can't inline it
-- away) and its declared sig must expand the cross-module synonym so `cam`
-- lowers to a Go struct (`cam.Center.X`), not box to
-- `cam.(map[string]any)["center"].(map[string]any)["x"]`.
liftCamera :: Camera -> Int
liftCamera cam = cam.center.x * cam.zoom + cam.center.y

sumCams :: List Camera -> Int
sumCams = case _ of
  Nil -> 0
  Cons c rest -> liftCamera c + sumCams rest

build :: Int -> List Camera
build n = case n of
  0 -> Nil
  _ -> Cons { center: { x: n, y: n * 2 }, zoom: n } (build (n - 1))

test :: Int
test = sumCams (build 3)
