package parser

import (
	"fmt"
	"strings"
	herr "hike/error"
	spc "hike/spec"
	tok "hike/token"
	hlv "hike/hilevel"
	loc "hike/location"
	abs "hike/abstract"
)

// ---------------------------------------- BuildFrame ----------------------------------------

type ParseFrame struct {
	What string
	Start *loc.Location
}

func (frame *ParseFrame) PrintErrorFrame(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Printf("parsing %s starting at ", frame.What)
	prn.Location(frame.Start)
	return prn.Done()
}

var _ herr.BuildFrame = &ParseFrame{}

// ---------------------------------------- BuildError ----------------------------------------

type SyntaxError struct {
	herr.BuildErrorBase
	Near *tok.Token
	Expected string
}

func (syntax *SyntaxError) PrintBuildError(level uint) error {
	prn := herr.NewErrorPrinter()
	prn.Level(level)
	prn.Print("Syntax error at ")
	prn.Location(&syntax.Near.Location)
	prn.Println()
	prn.Indent(1)
	prn.Inject(func(innerLevel uint) error {
		str, err := syntax.Near.Reconstruct()
		if err != nil {
			return err
		}
		prn.Print("found: ", str)
		return nil
	}, 0)
	prn.Println()
	prn.Indent(1)
	prn.Print("expected: ", syntax.Expected)
	syntax.InjectBacktrace(prn, 0)
	return prn.Done()
}

func (syntax *SyntaxError) BuildErrorLocation() *loc.Location {
	return &syntax.Near.Location
}

var _ herr.BuildError = &SyntaxError{}

// ---------------------------------------- Parser ----------------------------------------

type Parser struct {
	lexer chan *tok.Token
	Token *tok.Token
	firstError herr.BuildError
	knownStructures *KnownStructures
	specState *spc.State
}

type TopParser func(parser *Parser)
type ActionParser func(parser *Parser) abs.Action
type ArtifactParser func(parser *Parser) abs.Artifact
type TransformParser func(parser *Parser) abs.Transform
type TransformFactoryParser func(parser *Parser) hlv.TransformFactory
type ArtifactFactoryParser func(parser *Parser) hlv.ArtifactFactory
type ArtifactSetParser func(parser *Parser) []abs.Artifact
type FileFilterParser func(parser *Parser) hlv.FileFilter

type KnownStructures struct {
	top map[string]TopParser
	action map[string]ActionParser
	artifact map[string]ArtifactParser
	transform map[string]TransformParser
	transformFactory map[string]TransformFactoryParser
	artifactFactory map[string]ArtifactFactoryParser
	artifactSet map[string]ArtifactSetParser
	fileFilter map[string]FileFilterParser
}

func NewKnownStructures() *KnownStructures {
	return &KnownStructures {
		top: make(map[string]TopParser),
		action: make(map[string]ActionParser),
		artifact: make(map[string]ArtifactParser),
		transform: make(map[string]TransformParser),
		transformFactory: make(map[string]TransformFactoryParser),
		artifactFactory: make(map[string]ArtifactFactoryParser),
		artifactSet: make(map[string]ArtifactSetParser),
		fileFilter: make(map[string]FileFilterParser),
	}
}

func (known *KnownStructures) RegisterTopParser(initiator string, parser TopParser) {
	known.top[initiator] = parser
}

func (known *KnownStructures) TopParser(initiator string) TopParser {
	return known.top[initiator]
}

func (known *KnownStructures) RegisterActionParser(initiator string, parser ActionParser) {
	known.action[initiator] = parser
}

func (known *KnownStructures) ActionParser(initiator string) ActionParser {
	return known.action[initiator]
}

func (known *KnownStructures) RegisterArtifactParser(initiator string, parser ArtifactParser) {
	known.artifact[initiator] = parser
}

func (known *KnownStructures) ArtifactParser(initiator string) ArtifactParser {
	return known.artifact[initiator]
}

func (known *KnownStructures) RegisterTransformParser(initiator string, parser TransformParser) {
	known.transform[initiator] = parser
}

func (known *KnownStructures) TransformParser(initiator string) TransformParser {
	return known.transform[initiator]
}

func (known *KnownStructures) RegisterTransformFactoryParser(initiator string, parser TransformFactoryParser) {
	known.transformFactory[initiator] = parser
}

func (known *KnownStructures) TransformFactoryParser(initiator string) TransformFactoryParser {
	return known.transformFactory[initiator]
}

func (known *KnownStructures) RegisterArtifactFactoryParser(initiator string, parser ArtifactFactoryParser) {
	known.artifactFactory[initiator] = parser
}

func (known *KnownStructures) ArtifactFactoryParser(initiator string) ArtifactFactoryParser {
	return known.artifactFactory[initiator]
}

func (known *KnownStructures) RegisterArtifactSetParser(initiator string, parser ArtifactSetParser) {
	known.artifactSet[initiator] = parser
}

func (known *KnownStructures) ArtifactSetParser(initiator string) ArtifactSetParser {
	return known.artifactSet[initiator]
}

func (known *KnownStructures) RegisterFileFilterParser(initiator string, parser FileFilterParser) {
	known.fileFilter[initiator] = parser
}

func (known *KnownStructures) FileFilterParser(initiator string) FileFilterParser {
	return known.fileFilter[initiator]
}

func New(
	lexer chan *tok.Token,
	knownStructures *KnownStructures,
	specState *spc.State,
) *Parser {
	return &Parser {
		lexer: lexer,
		Token: <-lexer,
		knownStructures: knownStructures,
		specState: specState,
	}
}

func (parser *Parser) SpecState() *spc.State {
	return parser.specState
}

func (parser *Parser) Next() {
	if parser.Token.Type != tok.T_EOF {
		parser.Token = <-parser.lexer
	}
}

func (parser *Parser) Drain() {
	for parser.Token.Type != tok.T_EOF {
		parser.Token = <-parser.lexer
	}
}

func (parser *Parser) Error() herr.BuildError {
	return parser.firstError
}

func (parser *Parser) IsError() bool {
	return parser.firstError != nil
}

func (parser *Parser) Die(expected string) {
	if parser.firstError == nil {
		parser.firstError = &SyntaxError {
			Near: parser.Token,
			Expected: expected,
		}
	}
}

func (parser *Parser) Expect(ttype tok.Type) bool {
	if parser.Token.Type == ttype {
		return true
	}
	parser.Die(tok.NameType(ttype))
	return false
}

func (parser *Parser) ExpectExp(ttype tok.Type, explanation string) bool {
	if parser.Token.Type == ttype {
		return true
	}
	parser.Die(fmt.Sprintf("%s (%s)", tok.NameType(ttype), explanation))
	return false
}

func (parser *Parser) ExpectKeyword(name string) bool {
	if parser.Token.Type == tok.T_NAME && parser.Token.Text == name {
		return true
	}
	parser.Die(fmt.Sprintf("'%s'", name))
	return false
}

func (parser *Parser) IsKeyword(name string) bool {
	return parser.Token.Type == tok.T_NAME && parser.Token.Text == name
}

func (parser *Parser) InterpolateString() string {
	return parser.specState.InterpolateString(parser.Token.Text)
}

func (parser *Parser) Fail(fault herr.BuildError) {
	if parser.firstError == nil {
		parser.firstError = fault
	}
}

func (parser *Parser) SupersedeError(fault herr.BuildError) {
	if fault != nil {
		parser.firstError = fault
	}
}

func (parser *Parser) Frame(what string, start *loc.Location) {
	if start == nil {
		start = &parser.Token.Location
	}
	if parser.firstError != nil {
		parser.firstError.AddErrorFrame(&ParseFrame {
			What: what,
			Start: start,
		})
	}
}

func (parser *Parser) PushLexer(newLexer chan *tok.Token) chan *tok.Token {
	oldLexer := parser.lexer
	parser.lexer = newLexer
	parser.Token = <-newLexer
	return oldLexer
}

func (parser *Parser) PopLexer(oldLexer chan *tok.Token) {
	parser.lexer = oldLexer
	parser.Token = <-oldLexer
}

func (parser *Parser) Top() {
	if !parser.Expect(tok.T_NAME) {
		return
	}
	cb := parser.knownStructures.TopParser(parser.Token.Text)
	if cb == nil {
		parser.Die("top-level definition")
	} else {
		cb(parser)
	}
}

func (parser *Parser) IsTop() bool {
	return parser.Token.Type == tok.T_NAME && parser.knownStructures.TopParser(parser.Token.Text) != nil
}

func (parser *Parser) Action() abs.Action {
	if !parser.Expect(tok.T_NAME) {
		return nil
	}
	cb := parser.knownStructures.ActionParser(parser.Token.Text)
	if cb == nil {
		parser.Die("action")
		return nil
	} else {
		return cb(parser)
	}
}

func (parser *Parser) IsAction() bool {
	return parser.Token.Type == tok.T_NAME && parser.knownStructures.ActionParser(parser.Token.Text) != nil
}

func (parser *Parser) Artifact() abs.Artifact {
	if !parser.Expect(tok.T_NAME) {
		return nil
	}
	cb := parser.knownStructures.ArtifactParser(parser.Token.Text)
	if cb == nil {
		parser.Die("artifact")
		return nil
	} else {
		return cb(parser)
	}
}

func (parser *Parser) IsArtifact() bool {
	return parser.Token.Type == tok.T_NAME && parser.knownStructures.ArtifactParser(parser.Token.Text) != nil
}

func (parser *Parser) Transform() abs.Transform {
	if !parser.Expect(tok.T_NAME) {
		return nil
	}
	cb := parser.knownStructures.TransformParser(parser.Token.Text)
	if cb == nil {
		parser.Die("transform")
		return nil
	} else {
		return cb(parser)
	}
}

func (parser *Parser) IsTransform() bool {
	return parser.Token.Type == tok.T_NAME && parser.knownStructures.TransformParser(parser.Token.Text) != nil
}

func (parser *Parser) TransformFactory() hlv.TransformFactory {
	if !parser.Expect(tok.T_NAME) {
		return nil
	}
	cb := parser.knownStructures.TransformFactoryParser(parser.Token.Text)
	if cb == nil {
		parser.Die("transform factory")
		return nil
	} else {
		return cb(parser)
	}
}

func (parser *Parser) IsTransformFactory() bool {
	return parser.Token.Type == tok.T_NAME && parser.knownStructures.TransformFactoryParser(parser.Token.Text) != nil
}

func (parser *Parser) ArtifactFactory() hlv.ArtifactFactory {
	if !parser.Expect(tok.T_NAME) {
		return nil
	}
	cb := parser.knownStructures.ArtifactFactoryParser(parser.Token.Text)
	if cb == nil {
		parser.Die("artifact factory")
		return nil
	} else {
		return cb(parser)
	}
}

func (parser *Parser) IsArtifactFactory() bool {
	return parser.Token.Type == tok.T_NAME && parser.knownStructures.ArtifactFactoryParser(parser.Token.Text) != nil
}

func (parser *Parser) ArtifactSet() []abs.Artifact {
	if !parser.Expect(tok.T_NAME) {
		return nil
	}
	single := parser.knownStructures.ArtifactParser(parser.Token.Text)
	if single != nil {
		artifact := single(parser)
		if artifact == nil {
			return nil
		}
		return []abs.Artifact{artifact}
	}
	multi := parser.knownStructures.ArtifactSetParser(parser.Token.Text)
	if multi == nil {
		parser.Die("artifact set")
		return nil
	} else {
		return multi(parser)
	}
}

func (parser *Parser) IsArtifactSet() bool {
	switch {
		case parser.Token.Type != tok.T_NAME:
			return false
		case parser.knownStructures.ArtifactParser(parser.Token.Text) != nil:
			return true
		case parser.knownStructures.ArtifactSetParser(parser.Token.Text) != nil:
			return true
		default:
			return false
	}
}

func (parser *Parser) FileFilter() hlv.FileFilter {
	if !parser.Expect(tok.T_NAME) {
		return nil
	}
	cb := parser.knownStructures.FileFilterParser(parser.Token.Text)
	if cb == nil {
		parser.Die("file filter")
		return nil
	} else {
		return cb(parser)
	}
}

func (parser *Parser) IsFileFilter() bool {
	return parser.Token.Type == tok.T_NAME && parser.knownStructures.FileFilterParser(parser.Token.Text) != nil
}

// ---------------------------------------- intrinsics ----------------------------------------

func (parser *Parser) Utterance() {
	for parser.firstError == nil && parser.Token.Type != tok.T_EOF {
		parser.Top()
	}
}

type ArtifactRef interface {
	InjectArtifact(specState *spc.State, injector func(abs.Artifact))
}

type PresentArtifactRef struct {
	Artifact abs.Artifact
}

func (ref *PresentArtifactRef) InjectArtifact(specState *spc.State, injector func(abs.Artifact)) {
	injector(ref.Artifact)
}

var _ ArtifactRef = &PresentArtifactRef{}

type PendingArtifactRef struct {
	Key *abs.ArtifactKey
	ReferenceLocation *loc.Location
	ReferenceArise *herr.AriseRef
}

func (ref *PendingArtifactRef) InjectArtifact(specState *spc.State, injector func(abs.Artifact)) {
	specState.SlateResolver(func() herr.BuildError {
		artifact := specState.Artifact(ref.Key)
		if artifact != nil {
			injector(artifact)
			return nil
		} else {
			return &spc.NoSuchArtifactError {
				Key: ref.Key,
				ReferenceLocation: ref.ReferenceLocation,
				ReferenceArise: ref.ReferenceArise,
			}
		}
	})
}

var _ ArtifactRef = &PendingArtifactRef{}

func SplitArtifactKey(ks string, config *spc.Config) *abs.ArtifactKey {
	key := &abs.ArtifactKey{}
	pos := strings.Index(ks, "::")
	if pos < 0 {
		key.Project = config.EffectiveProjectName()
		key.Artifact = ks
	} else {
		key.Project = ks[:pos]
		key.Artifact = ks[pos + 2:]
	}
	return key
}

func (parser *Parser) artifactRefHead(extended bool) bool {
	if extended {
		if parser.Token.Type != tok.T_NAME || parser.Token.Text != "artifact" {
			return false
		}
		start := &parser.Token.Location
		parser.Next()
		if !parser.ExpectExp(tok.T_STRING, "referenced artifact key") {
			parser.Frame("artifact reference", start)
		}
		return true
	} else {
		return parser.Token.Type == tok.T_STRING
	}
}

func (parser *Parser) ArtifactRef(arise *herr.AriseRef, extended bool) ArtifactRef {
	switch {
		case parser.artifactRefHead(extended):
			if parser.IsError() {
				return nil
			}
			specState := parser.SpecState()
			key := SplitArtifactKey(parser.InterpolateString(), specState.Config)
			refLocation := &parser.Token.Location
			parser.Next()
			artifact := specState.Artifact(key)
			if artifact != nil {
				return &PresentArtifactRef {
					Artifact: artifact,
				}
			} else {
				return &PendingArtifactRef {
					Key: key,
					ReferenceLocation: refLocation,
					ReferenceArise: arise,
				}
			}
		case parser.IsArtifact():
			artifact := parser.Artifact()
			if artifact == nil {
				return nil
			}
			return &PresentArtifactRef {
				Artifact: artifact,
			}
		default:
			parser.Die("string (artifact key) or artifact")
			return nil
	}
}

func (parser *Parser) IsArtifactRef(extended bool) bool {
	if parser.IsArtifact() {
		return true
	}
	if extended {
		return parser.IsKeyword("artifact")
	} else {
		return parser.Token.Type == tok.T_STRING
	}
}
