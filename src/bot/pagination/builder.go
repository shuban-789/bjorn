package pagination

type PaginationBuilder struct {
	Paginator *Paginator
}

func New(CustomIdPrefix string) *PaginationBuilder {
	return &PaginationBuilder{
		Paginator: &Paginator{
			CustomIDPrefix: CustomIdPrefix,
			ExtraDataKeys:  []string{},
		},
	}
}

func (pb *PaginationBuilder) AddExtraKey(key string) *PaginationBuilder {
	pb.Paginator.ExtraDataKeys = append(pb.Paginator.ExtraDataKeys, key)
	return pb
}

func (pb *PaginationBuilder) OnCreate(createFunc CreatePage) *PaginationBuilder {
	pb.Paginator.Create = createFunc
	return pb
}

func (pb *PaginationBuilder) OnUpdate(updateFunc UpdatePage) *PaginationBuilder {
	pb.Paginator.Update = updateFunc
	return pb
}

func (pb *PaginationBuilder) Register() *Paginator {
	pb.Paginator.Register()
	return pb.Paginator
}
