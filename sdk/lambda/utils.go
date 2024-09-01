package sdklambda

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/aws/aws-lambda-go/events"
)

func UnmarshalAs[T any](bts []byte, out T) (T, error) {
	decoder := json.NewDecoder(bytes.NewBuffer(bts))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&out)
	return out, err
}

type OwnEventHandler func(context.Context, []byte) (bool, any, error)

func GetOwnEventHandler[E any, R any](
	evt E,
	onOwnEvent func(ctx context.Context, evt E) (R, error),
) OwnEventHandler {
	return func(ctx context.Context, b []byte) (bool, any, error) {
		if ownEvt, err := UnmarshalAs(b, evt); err == nil {
			r, e := onOwnEvent(ctx, ownEvt)
			return true, r, e
		}
		return false, nil, nil
	}
}

func MultiEventTypeHandler(
	handlers []OwnEventHandler,
) func(context.Context, json.RawMessage) (any, error) {

	return func(ctx context.Context, rm json.RawMessage) (any, error) {

		fHandled := false
		var fRes any
		var err error

		wg := sync.WaitGroup{}
		mw := sync.Mutex{}

		for _, h := range handlers {
			wg.Add(1)
			go func(h OwnEventHandler) {
				defer wg.Done()
				handled, res, lerr := h(ctx, rm)
				if err != nil {
					log.Println("error", "while processing via handler", err, h, string(rm))
				}
				if handled {
					mw.Lock()
					defer mw.Unlock()

					if !fHandled {
						fHandled = true
						fRes = res
						err = lerr
					} else {
						log.Println("warn", "multiple handlers are processing this event", h, string(rm))
					}
				}
			}(h)
		}

		wg.Wait()

		if fHandled {
			return fRes, err
		}

		log.Println("error", "unable to handle request", string(rm))
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
		}, errors.New("unable to handle request type")
	}
}
