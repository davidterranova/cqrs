# CQRS
Generic Eventsourcing / CQRS implementation in Go with a focus on simplicity

## Write model: creating a new aggregate

1. Define a new struct that embed `eventsourcing.AggregateBase`
   ```go
   const GroupAggregateType eventsourcing.AggregateType = "group"

   type Group struct {
     *eventsourcing.AggregateBase[Group]

     name string
   }

   func NewGroup() *Group {
      return &Group{
        AggregateBase: eventsourcing.NewAggregateBase[Group](uuid.Nil, 0),
      }
    }

   func (g Group) AggregateType() eventsourcing.AggregateType {
     return GroupAggregateType
   }
   ```

2. Define and register events
   ```go
   const EvtTypeGroupCreated       eventsourcing.EventType = "group.created"

   func RegisterGroupEvents(registry eventsourcing.EventRegistry[Group]) {
      registry.Register(EvtTypeGroupCreated, func() eventsourcing.Event[Group] {
        return &EvtGroupCreated{
          EventBase: &eventsourcing.EventBase[Group]{},
        }
      })
    }

    type EvtGroupCreated struct {
      *eventsourcing.EventBase[Group]
    }

    func NewEvtGroupCreated(aggregateId uuid.UUID, aggregateVersion int, createdBy eventsourcing.User) *EvtGroupCreated {
      return &EvtGroupCreated{
        EventBase: eventsourcing.NewEventBase[Group](
          GroupAggregateType,
          aggregateVersion,
          EvtTypeGroupCreated,
          aggregateId,
          createdBy,
        ),
      }
    }

    func (e EvtGroupCreated) Apply(g *Group) error {
      g.Init(e)

      return nil
    }
   ```
3. Create use case
   ```go
   type CmdCreate struct {
      Name string
    }

    type cmdCreate struct {
      eventsourcing.CommandBase[domain.Group]
      CmdCreate
    }

    type CreateHandler struct {
      validator      *validator.Validate
      commandHandler eventsourcing.CommandHandler[domain.Group]
    }

    func NewCreateHandler(commandHandler eventsourcing.CommandHandler[domain.Group]) CreateHandler {
      validator := validator.New(validator.WithRequiredStructEnabled())

      return CreateHandler{
        commandHandler: commandHandler,
        validator:      validator,
      }
    }

    func (h CreateHandler) Create(ctx context.Context, issuer eventsourcing.User, cmd CmdCreate) (*domain.Group, error) {
      err := h.Validate(cmd)
      if err != nil {
        log.Error().Err(err).Msg("create: invalid command")
        return nil, err
      }

      // check if possible to create a group for this issuer

      internalCmd := cmdCreate{
        CommandBase: eventsourcing.NewCommandBase[domain.Group](
          uuid.New(),
          domain.AggregateGroup,
          issuer,
        ),
        CmdCreate: cmd,
      }

      return h.commandHandler.HandleCommand(ctx, internalCmd)
    }

    func (h CreateHandler) Validate(cmd CmdCreate) error {
      return h.validator.Struct(cmd)
    }

    func (c cmdCreate) Apply(aggregate *domain.Group) ([]eventsourcing.Event[domain.Group], error) {
      err := eventsourcing.EnsureNewAggregate(aggregate)
      if err != nil {
        return nil, fmt.Errorf("create group: %w", err)
      }

      events := []eventsourcing.Event[domain.Group]{
        domain.NewEvtGroupCreated(c.AggregateId(), 0, c.IssuedBy()),
        domain.NewEvtGroupNameSet(c.AggregateId(), 1, c.IssuedBy(), c.Name),
      }

      return events, nil
    }
   ```

## Read model: handling events

1. Create a new read model (in memory for the example, relying on the generic one from the library)
   ```go
    type GroupQuery struct {
      id *uuid.UUID
    }

    func NewGroupQuery(id *uuid.UUID) *GroupQuery {
      return &GroupQuery{
        id: id,
      }
    }

    func (q GroupQuery) Id() *uuid.UUID {
      return q.id
    }

    type inMemoryReadModel struct {
      rm *readmodel.InMemoryReadModel[domain.Group]
    }

    func NewInMemoryReadModel(eventStream eventsourcing.Subscriber[domain.Group]) *inMemoryReadModel {
      return &inMemoryReadModel{
        rm: readmodel.NewInMemoryReadModel[domain.Group](
          eventStream,
          domain.NewGroup,
          domain.EvtTypeGroupCreated,
          eventsourcing.EvtTypeNil, // Replace by the appropriate event type when the event is implemented
          eventsourcing.EvtTypeNil, // Replace by the appropriate events type when the event is implemented. Make sure to list all the events the read model should be handling
        ),
      }
    }

    // HandleEvent definition is optional as Find and Get will handle the job
    // however it might be useful to have it exposed for testing purpose (manually injecting events)
    func (l *inMemoryReadModel) HandleEvent(e eventsourcing.Event[domain.Group]) {
      l.rm.HandleEvent(e)
    }

    func (l *inMemoryReadModel) Find(ctx context.Context, query GroupQuery) ([]*domain.Group, error) {
      return l.rm.Find(ctx, aggregateMatcher(query))
    }

    func (l *inMemoryReadModel) Get(ctx context.Context, query GroupQuery) (*domain.Group, error) {
      return l.rm.Get(ctx, aggregateMatcher(query))
    }

    func aggregateMatcher(query GroupQuery) readmodel.AggregateMatcher[domain.Group] {
      var matcher readmodel.AggregateMatcher[domain.Group]

      if query != nil {
        matcher = readmodel.AggregateMatcherAnd[domain.Group](
          readmodel.AggregateMatcherAggregateId[domain.Group](query.Id()),
        )
      }

      return matcher
    }
   ```
2. Create the list use case
   ```go
   type GroupReadModel interface {
      Find(ctx context.Context, query GroupQuery) ([]*domain.Group, error)
      Get(ctx context.Context, query GroupQuery) (*domain.Group, error)
    }

    type ListHandler struct {
      groupReadModel GroupReadModel
    }

    func NewListHandler(groupReadModel GroupReadModel) ListHandler {
      return ListHandler{
        groupReadModel: groupReadModel,
      }
    }

    func (h ListHandler) List(ctx context.Context, issuedBy eventsourcing.User, query GroupQuery) ([]*domain.Group, error) {
      // TODO: check if possible to list groups for this issuer
      return h.groupReadModel.Find(ctx, query)
    }
   ```

![event sourcing approach](./doc/eventsourcing.png)
