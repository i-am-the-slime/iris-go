module Main where

import Prelude

import Control.Alt ((<|>))
import Data.Array as Array
import Data.Either (Either(..))
import Data.Enum (fromEnum)
import Data.Foldable (for_)
import Data.Maybe (Maybe(..), fromMaybe, maybe)
import Data.Maybe (optional) as Maybe
import Data.Number as Number
import Data.String as String
import Data.Tuple.Nested (type (/\), (/\))
import Effect (Effect)
import Effect.Console (log)
import ExitCodes as ExitCode
import Options.Applicative as Opts
import Text.PrettyPrint.Leijen as PP

data Command
  = Render RenderArgs
  | Check { files :: Array String, colour :: Boolean }
  | Serve { port :: Int, host :: String }

type RenderArgs =
  { input :: Maybe String
  , output :: String
  , fps :: Int
  , scale :: Number
  , theme :: Maybe Theme
  , quiet :: Boolean
  , tags :: Array String
  }

data Theme = Light | Dark | Blueprint

type Global = { verbose :: Boolean, config :: Maybe String }

main :: Effect Unit
main = for_ invocations \arguments -> do
  log ("$ markgraf " <> String.joinWith " " arguments)
  log (run arguments)
  where
  invocations =
    [ [ "render", "hello.markgraf", "-o", "out.gif", "--fps", "30", "--scale", "1.5", "--theme", "dark", "--tag", "a", "--tag", "b" ]
    , [ "--verbose", "--config", "cfg.json", "check", "a.markgraf", "b.markgraf", "--colour" ]
    , [ "check", "--color" ]
    , [ "serve", "--port", "8080" ]
    , [ "serve" ]
    , [ "render" ]
    , [ "render", "--theme", "neon" ]
    , [ "render", "--fps", "fast" ]
    , [ "launch" ]
    , [ "--version" ]
    , [ "--help" ]
    , [ "render", "--help" ]
    , []
    ]

run :: Array String -> String
run arguments = case Opts.execParserPure Opts.defaultPrefs cli arguments of
  Opts.Success (global /\ command) -> describeGlobal global <> "\n" <> describe command
  Opts.Failure failure -> do
    let text /\ code = Opts.renderFailure failure "markgraf"
    "exit " <> show (fromEnum code) <> exitKind code <> "\n" <> text
  Opts.CompletionInvoked _ -> "completion"
  where
  exitKind code
    | code == ExitCode.Success = " (success)"
    | otherwise = " (error)"

cli :: Opts.ParserInfo (Global /\ Command)
cli = Opts.info (programParser Opts.<**> versionOption Opts.<**> Opts.helper)
  ( Opts.fullDesc
      <> Opts.progDescDoc (Just exampleDoc)
      <> Opts.header "markgraf - render markgraf animations"
      <> Opts.footer "See https://example.com/markgraf for more."
  )
  where
  programParser = (/\) <$> globalParser <*> commandParser
  versionOption = Opts.infoOption "markgraf 1.2.3"
    ( Opts.long "version"
        <> Opts.short 'V'
        <> Opts.help "Print version and exit"
    )

exampleDoc :: PP.Doc
exampleDoc = PP.string $ String.joinWith "\n"
  [ "Render, check or serve .markgraf diagrams."
  , ""
  , "Example:  markgraf render hello.markgraf -o hello.mp4"
  ]

globalParser :: Opts.Parser Global
globalParser = ado
  verbose <- Opts.switch (Opts.long "verbose" <> Opts.help "Log every step")
  config <- Maybe.optional
    ( Opts.strOption
        ( Opts.long "config"
            <> Opts.short 'c'
            <> Opts.metavar "FILE"
            <> Opts.help "Read settings from FILE"
        )
    )
  in { verbose, config }

commandParser :: Opts.Parser Command
commandParser = Opts.hsubparser
  ( Opts.command "render" (Opts.info renderParser (Opts.progDesc "Render a diagram to video"))
      <> Opts.command "check" (Opts.info checkParser (Opts.progDesc "Validate diagrams"))
      <> Opts.command "serve" (Opts.info serveParser (Opts.progDesc "Serve frames over HTTP"))
  )

renderParser :: Opts.Parser Command
renderParser = map Render ado
  input <- Maybe.optional (Opts.strArgument (Opts.metavar "INPUT" <> Opts.help "Source file (omit for stdin)"))
  output <- Opts.strOption
    ( Opts.long "output"
        <> Opts.short 'o'
        <> Opts.metavar "PATH"
        <> Opts.value "out.mp4"
        <> Opts.showDefault
        <> Opts.help "Output video path"
    )
  fps <- Opts.option Opts.int (Opts.long "fps" <> Opts.metavar "INT" <> Opts.value 60 <> Opts.help "Frames per second")
  scale <- Opts.option (Opts.eitherReader finiteNumber) (Opts.long "scale" <> Opts.metavar "NUMBER" <> Opts.value 2.0 <> Opts.help "Resolution multiplier")
  theme <- Maybe.optional (Opts.option (Opts.eitherReader themeReader) (Opts.long "theme" <> Opts.metavar "NAME" <> Opts.help "light | dark | blueprint"))
  quiet <- Opts.switch (Opts.long "quiet" <> Opts.short 'q' <> Opts.help "Suppress progress output")
  tags <- Array.fromFoldable <$> Opts.many (Opts.strOption (Opts.long "tag" <> Opts.metavar "TAG" <> Opts.help "Attach a tag (repeatable)"))
  in { input, output, fps, scale, theme, quiet, tags }

checkParser :: Opts.Parser Command
checkParser = ado
  files <- Array.fromFoldable <$> Opts.many (Opts.argument Opts.str (Opts.metavar "FILES..."))
  colour <- Opts.switch (Opts.long "colour" <> Opts.help "Colour diagnostics") <|> Opts.switch (Opts.long "color" <> Opts.internal)
  in Check { files, colour }

serveParser :: Opts.Parser Command
serveParser = ado
  port <- Opts.option Opts.int (Opts.long "port" <> Opts.short 'p' <> Opts.metavar "PORT" <> Opts.value 4000 <> Opts.showDefault <> Opts.help "Port to listen on")
  host <- Opts.strOption (Opts.long "host" <> Opts.metavar "HOST" <> Opts.help "Interface to bind") <|> pure "localhost"
  in Serve { port, host }

finiteNumber :: String -> Either String Number
finiteNumber text = case Number.fromString text of
  Just number | number > 0.0 -> Right number
  _ -> Left ("expected a positive number, got " <> show text)

themeReader :: String -> Either String Theme
themeReader = case _ of
  "light" -> Right Light
  "dark" -> Right Dark
  "blueprint" -> Right Blueprint
  other -> Left ("unknown theme " <> show other)

describeGlobal :: Global -> String
describeGlobal global = "verbose=" <> show global.verbose <> " config=" <> fromMaybe "-" global.config

describe :: Command -> String
describe = case _ of
  Render args -> String.joinWith " "
    [ "render"
    , "input=" <> fromMaybe "<stdin>" args.input
    , "output=" <> args.output
    , "fps=" <> show args.fps
    , "scale=" <> show args.scale
    , "theme=" <> maybe "default" themeName args.theme
    , "quiet=" <> show args.quiet
    , "tags=" <> show args.tags
    ]
  Check args -> "check files=" <> show args.files <> " colour=" <> show args.colour
  Serve args -> "serve " <> args.host <> ":" <> show args.port
  where
  themeName = case _ of
    Light -> "light"
    Dark -> "dark"
    Blueprint -> "blueprint"
