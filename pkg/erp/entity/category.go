package entity

type Category struct {
	// Source consists of the cluster id and namespace in the form of "zone:namespace".
	Source string
	// Target contains a unique identifier of a Category representation in the foreign ERP.
	Target string
}
