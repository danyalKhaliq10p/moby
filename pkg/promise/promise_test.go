package promise

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGo(t *testing.T) {
	errCh := Go(functionWithError)
	er := <-errCh
	require.NotNil(t, er.Error())
	require.EqualValues(t, "Error Occured", er.Error())
	t.Log("GO promise with channel value 'Error' executed successfully")

	noErrCh := Go(functionWithNoError)
	er = <-noErrCh
	require.Nil(t, er)
	t.Log("GO promise with channel value 'No Error' executed successfully")
}

func functionWithError() (err error) {
	return errors.New("Error Occured")
}
func functionWithNoError() (err error) {
	return nil
}
