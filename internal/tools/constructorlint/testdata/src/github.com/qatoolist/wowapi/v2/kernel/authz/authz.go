package authz

type Store struct{}

func NewStore() *Store { return &Store{} }

func NewSQLStore() *Store { return &Store{} }
