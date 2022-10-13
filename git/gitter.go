package git

import "context"

type Gitter interface {
	Clone(context.Context, *CloneOptions) error
}
