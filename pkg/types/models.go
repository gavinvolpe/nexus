package types

type IRunOutput interface {
	Execute(tmpl string) (string, error)
	Callback() error
}

type IFlow interface {
	Flow(...IScenario) IAction
	Run() (IRunOutput, error)
}

type ISequence interface {
	Sequence(...IAction) IAction
	Run() (IRunOutput, error)
}

type IParallel interface {
	Parallel(...IAction) IAction
	Run() (IRunOutput, error)
}

type IExpression interface {
	Expression(string) IExpression
	Run() (IRunOutput, error)
}

type IConcurrent interface {
	Actions(...IAction) IAction
	Scenarios(...IScenario) IScenario
	Sequences(...ISequence) ISequence
	Expressions(...IExpression) IExpression
	Run() (IRunOutput, error)
}

type IScenario interface {
	When(...IScenario) IScenario
	Then(...IScenario) IScenario
	Else(...IScenario) IScenario
	Run() (IRunOutput, error)
}

type IAction interface {
	Call() IOutput
	Sequence(...IAction) IAction
	Concurrent(...IAction) IAction
	Parallel(...IAction) IAction
	Flow(...IScenario) IAction
	Run() (IRunOutput, error)
}

type IOutput interface {
	Parse(string) []IAction
}

// PromptSystem interfaces

type IPrompt interface {
	GetContent() string
	GetVariables() map[string]string
	GetPurpose() string
	GetTarget() string
	GetWhen() string
	Validate() error
}

type IPromptTemplate interface {
	Render(vars map[string]string) (string, error)
	Parse(content string) error
	GetTemplateType() string
}

type IPipelineStage interface {
	Process(input interface{}) (interface{}, error)
	GetStageType() string
	GetConfig() map[string]interface{}
	Validate() error
}

type IAgent interface {
	ID() string
	GetTools() []ITool
	GetCapabilities() []string
	ProcessPrompt(prompt IPrompt) (IActionResult, error)
	ExecuteAction(action IAction) (IActionResult, error)
}

type ITool interface {
	Name() string
	Description() string
	Execute(params map[string]interface{}) (interface{}, error)
	ValidateParams(params map[string]interface{}) error
}

type IActionResult interface {
	GetOutput() interface{}
	GetError() error
	GetMetadata() map[string]interface{}
	IsSuccess() bool
}

type IPipeline interface {
	AddStage(stage IPipelineStage) error
	RemoveStage(stageID string) error
	Execute(input interface{}) (interface{}, error)
	GetStages() []IPipelineStage
	Validate() error
}

type IPromptSelector interface {
	SelectPrompt(scenario string, context map[string]interface{}) (IPrompt, error)
	AddScenario(scenario string, prompt IPrompt) error
	RemoveScenario(scenario string) error
}

type IPromptManager interface {
	LoadPrompts(source string) error
	GetPrompt(id string) (IPrompt, error)
	GetPromptsByTarget(target string) []IPrompt
	GetPromptsByWhen(when string) []IPrompt
	ValidatePrompts() []error
}

type IPipelineBuilder interface {
	CreatePipeline() IPipeline
	AddProcessingStage(processor IPipelineStage) IPipelineBuilder
	AddAgent(agent IAgent) IPipelineBuilder
	AddOutputHandler(handler func(interface{}) error) IPipelineBuilder
	Build() (IPipeline, error)
}

type IWorkflow interface {
	ID() string
	AddStep(step IPipeline) error
	RemoveStep(stepID string) error
	Execute() ([]IActionResult, error)
	GetStatus() string
	Validate() error
}

type IContext interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	GetAll() map[string]interface{}
	Clear()
}

// AccessPattern interfaces

type IAccessNode interface {
	ID() string
	GetType() string
	GetMetadata() map[string]interface{}
	GetRequirements() []string
	GetCapabilities() []string
	ValidateNode() error
}

type IAccessEdge interface {
	From() IAccessNode
	To() IAccessNode
	Weight() float64
	GetTransformation() ITransformation
	GetConstraints() []IConstraint
	ValidateEdge() error
}

type ITransformation interface {
	Transform(input interface{}) (interface{}, error)
	GetType() string
	GetCost() float64
	ValidateTransform() error
}

type IConstraint interface {
	Evaluate(context IContext) bool
	GetPriority() int
	GetDescription() string
}

type IKnowledgeGraph interface {
	AddNode(node IAccessNode) error
	AddEdge(edge IAccessEdge) error
	RemoveNode(nodeID string) error
	RemoveEdge(fromID, toID string) error
	FindPath(from, to IAccessNode) ([]IAccessEdge, error)
	ValidateGraph() error
}

type IPathStrategy interface {
	CalculatePath(graph IKnowledgeGraph, start, end IAccessNode) ([]IAccessEdge, error)
	GetCost(path []IAccessEdge) float64
	OptimizePath(path []IAccessEdge) ([]IAccessEdge, error)
}

type IAccessPattern interface {
	// Core pattern management
	ID() string
	GetName() string
	GetDescription() string
	GetVersion() string
	
	// Knowledge structure
	GetKnowledgeGraph() IKnowledgeGraph
	GetPathStrategy() IPathStrategy
	
	// Pattern execution
	Execute(input interface{}, context IContext) (interface{}, error)
	Validate() error
	
	// Pattern composition
	ComposeWith(other IAccessPattern) (IAccessPattern, error)
	DecomposeInto(subpatterns []IAccessPattern) error
	
	// Pattern evolution
	Learn(feedback IActionResult) error
	Adapt(context IContext) error
	
	// Integration with other interfaces
	GetRequiredAgents() []IAgent
	GetRequiredTools() []ITool
	CreatePipeline() (IPipeline, error)
	
	// Pattern analysis
	AnalyzeEfficiency() map[string]float64
	GetSuccessMetrics() map[string]interface{}
	SuggestOptimizations() []string
}

type IPatternRegistry interface {
	Register(pattern IAccessPattern) error
	Unregister(patternID string) error
	GetPattern(patternID string) (IAccessPattern, error)
	FindPatternsByCapability(capability string) []IAccessPattern
	FindPatternsByContext(context IContext) []IAccessPattern
	ValidateRegistry() error
}

type IPatternMatcher interface {
	Match(input interface{}, context IContext) ([]IAccessPattern, error)
	RankPatterns(patterns []IAccessPattern, context IContext) ([]IAccessPattern, error)
	ValidateMatcher() error
}

type IPatternComposer interface {
	Compose(patterns []IAccessPattern) (IAccessPattern, error)
	Decompose(pattern IAccessPattern) ([]IAccessPattern, error)
	OptimizeComposition(pattern IAccessPattern) (IAccessPattern, error)
	ValidateComposition(pattern IAccessPattern) error
}

type IPatternExecutor interface {
	PrepareExecution(pattern IAccessPattern, input interface{}) error
	Execute(context IContext) (IActionResult, error)
	MonitorExecution() map[string]interface{}
	HandleFailure(error) (IActionResult, error)
}

type IPatternLearner interface {
	LearnFromExecution(pattern IAccessPattern, result IActionResult) error
	UpdateWeights(graph IKnowledgeGraph, feedback map[string]float64) error
	GenerateInsights() []string
	ExportLearnings() map[string]interface{}
}
