package pagination

type PaginationBuilder[T any] struct {
	Paginator *Paginator[T]
}

func New[T any](CustomIdPrefix string) *PaginationBuilder[T] {
	return &PaginationBuilder[T]{
		Paginator: &Paginator[T]{
			CustomIDPrefix: CustomIdPrefix,
			ExtraDataKeys:  []string{},
		},
	}
}

func (pb *PaginationBuilder[T]) ItemsPerPage(count int) *PaginationBuilder[T] {
	pb.Paginator.ItemsPerPage = count
	return pb
}

func (pb *PaginationBuilder[T]) AddExtraKey(key string) *PaginationBuilder[T] {
	pb.Paginator.ExtraDataKeys = append(pb.Paginator.ExtraDataKeys, key)
	return pb
}

func (pb *PaginationBuilder[T]) OnCreate(createFunc CreatePage[T]) *PaginationBuilder[T] {
	pb.Paginator.Create = createFunc
	return pb
}

func (pb *PaginationBuilder[T]) OnUpdate(updateFunc UpdatePage[T]) *PaginationBuilder[T] {
	pb.Paginator.Update = updateFunc
	return pb
}

func (pb *PaginationBuilder[T]) WithDataGetter(getter func(state PaginationState) ([]T, error)) *PaginationBuilder[T] {
	pb.Paginator.GetData = getter
	return pb
}

func (pb *PaginationBuilder[T]) Register() *Paginator[T] {
	if pb.Paginator.Update == nil {
		panic("Paginator.Update is required")
	}
	if pb.Paginator.CustomIDPrefix == "" {
		panic("Paginator.CustomIDPrefix is required")
	}

	pb.Paginator.Register()
	return pb.Paginator
}
