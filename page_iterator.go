package msgraphgocore

import (
	"log"
	"net/url"
	"reflect"
	"unsafe"

	abstractions "github.com/microsoft/kiota/abstractions/go"
	"github.com/microsoft/kiota/abstractions/go/serialization"
	jsonserialization "github.com/microsoft/kiota/serialization/go/json"
)

type Page interface {
	getValue() []interface{}
	getNextLink() *string
}

type PageIterator struct {
	currentPage     Page
	reqAdapter      GraphRequestAdapterBase
	pauseIndex      int
	constructorFunc ParsableConstructor
	headers         map[string]string
}

type ParsableConstructor func() serialization.Parsable

type PageResult struct {
	nextLink *string
	value    []interface{}
}

func (p *PageResult) getValue() []interface{} {
	if p == nil {
		return nil
	}

	return p.value
}

func (p *PageResult) getNextLink() *string {
	if p == nil {
		return nil
	}

	return p.nextLink
}

func NewPageIterator(res interface{}, reqAdapter GraphRequestAdapterBase, constructorFunc ParsableConstructor, headers map[string]string) *PageIterator {
	abstractions.RegisterDefaultSerializer(func() serialization.SerializationWriterFactory {
		return jsonserialization.NewJsonSerializationWriterFactory()
	})
	abstractions.RegisterDefaultDeserializer(func() serialization.ParseNodeFactory {
		return jsonserialization.NewJsonParseNodeFactory()
	})

	return &PageIterator{
		convertToPage(res),
		reqAdapter,
		0, // pauseIndex helps us remember where we paused enumeration in the page.
		constructorFunc,
		headers,
	}
}

func (pI *PageIterator) Iterate(callback func(pageItem interface{}) bool) {
	for pI.currentPage != nil {
		keepIterating := pI.enumerate(callback)

		if !keepIterating {
			// Callback returned false, stop iterating through pages.
			return
		}

		pI.next()
		pI.pauseIndex = 0 // when moving to the next page reset pauseIndex
	}
}

func (pI *PageIterator) hasNext() bool {
	if pI.currentPage == nil || pI.currentPage.getNextLink() == nil {
		return false
	}
	return true
}

func (pI *PageIterator) next() Page {
	nextPage := pI.getNextPage()

	pI.currentPage = nextPage
	return nextPage
}

func (pI *PageIterator) getNextPage() *PageResult {
	if pI.currentPage.getNextLink() == nil {
		return nil
	}

	nextLink, err := url.Parse(*pI.currentPage.getNextLink())
	if err != nil {
		log.Fatal(err)
	}

	requestInfo := abstractions.NewRequestInformation()
	requestInfo.Method = abstractions.GET
	requestInfo.SetUri(*nextLink)
	requestInfo.Headers = pI.headers

	res, err := pI.reqAdapter.SendAsync(*requestInfo, pI.constructorFunc, nil)
	if err != nil {
		log.Fatal(err)
	}

	return convertToPage(res)
}

func (pI *PageIterator) enumerate(callback func(item interface{}) bool) bool {
	keepIterating := true

	if pI.currentPage == nil {
		return false
	}

	pageItems := pI.currentPage.getValue()
	if pageItems == nil {
		return false
	}

	if pI.pauseIndex >= len(pageItems) {
		return false
	}

	// start/continue enumerating page items from  pauseIndex.
	// this makes it possible to resume iteration from where we paused iteration.
	for i := pI.pauseIndex; i < len(pageItems); i++ {
		keepIterating = callback(pageItems[i])

		if !keepIterating {
			// Callback returned false, pause! stop enumerating page items. Set pauseIndex so that we know
			// where to resume from.
			// Resumes from the next item
			pI.pauseIndex = i + 1
			break
		}
	}

	return keepIterating
}

func convertToPage(response interface{}) *PageResult {
	ref := reflect.ValueOf(response).Elem()
	value := ref.FieldByName("value")
	value = reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem()

	nextLink := ref.FieldByName("nextLink")
	nextLink = reflect.NewAt(nextLink.Type(), unsafe.Pointer(nextLink.UnsafeAddr())).Elem()

	// Collect all entities in the value slice.
	// This converts a graph slice ie []graph.User to a dynamic slice []interface{}
	collected := make([]interface{}, 0)
	for i := 0; i < value.Len(); i++ {
		collected = append(collected, value.Index(i).Interface())
	}

	return &PageResult{
		nextLink: nextLink.Interface().(*string),
		value:    collected,
	}
}
