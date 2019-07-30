package transactionmetadata

import "github.com/iotaledger/goshimmer/packages/errors"

// region errors ///////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	ErrUnmarshalFailed = errors.Wrap(errors.New("unmarshall failed"), "input data is corrupted")
	ErrMarshallFailed  = errors.Wrap(errors.New("marshal failed"), "the source object contains invalid values")
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
