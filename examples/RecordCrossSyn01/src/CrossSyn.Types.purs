module CrossSyn.Types (Point, Camera) where

-- A record type synonym defined in a DIFFERENT module from the function that
-- consumes it. This is the markgraf `Camera`/`Placement` shape: a function in
-- Markgraf.Animation.Camera takes a param typed by `Camera`, a synonym from
-- Markgraf.Animation.State. The declared-sig alias expansion must resolve the
-- synonym by its QUALIFIED name to lower the param to a Go struct.
type Point = { x :: Int, y :: Int }

type Camera = { center :: Point, zoom :: Int }
