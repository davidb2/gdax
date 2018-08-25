package gdax

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// A pageable is an interface for an iterator.
type pageable interface {
	HasNext() bool
	Next() interface{}
}

// A pagination represents a cursor.
type pagination struct {
	before string
	after  string
	limit  int
}

// A pageableCollection represents an iterator of HTTP responses.
type pageableCollection struct {
	currentPage        int
	currentIndexInPage int
	accessInfo         *AccessInfo
	pagination
	size                  int
	finishedReadingPage   bool
	pendingError          error
	usesPaginationCursors bool
	pages                 [][]interface{}
}

// A DayHourMin is a time struct with format mm,hh,dd.
type DayHourMin struct {
	time.Time
}

// MarshalJSON creates JSON from a DayHourMin struct.
func (d *DayHourMin) MarshalJSON() ([]byte, error) {
	return []byte(d.Format("mm,hh,dd")), nil
}

// newPageableCollection creates a new pageable collection.
// Some pageable collections do not have HTTP paginations cursors (i.e., the HTTP response returns a single JSON array).
// In this case, "usesPaginationCursors" should be false.
func (accessInfo *AccessInfo) newPageableCollection(usesPaginationCursors bool) pageableCollection {
	return pageableCollection{
		currentPage:           -1,
		accessInfo:            accessInfo,
		pagination:            pagination{before: "", after: "", limit: -1},
		size:                  0,
		finishedReadingPage:   false,
		pendingError:          nil,
		usesPaginationCursors: usesPaginationCursors,
		pages: nil,
	}
}
func (p pagination) String() string {
	var (
		before string
		after  string
		limit  string
	)
	if p.before != "" {
		before = fmt.Sprintf("before=%s", p.before)
	}
	if p.after != "" {
		after = fmt.Sprintf("after=%s", p.after)
	}
	if p.limit != -1 {
		limit = fmt.Sprintf("limit=%d", p.limit)
	}
	return strings.Join(stringFilter([]string{before, after, limit}, notEmpty), "&")
}

// hasNext determines if the pageableCollection has any more elements.
// Also, this function loads the additional elements into memory if there are more elements.
func (c *pageableCollection) hasNext(method, path, params, body string, container interface{}) bool {
	if len(c.pages) == 0 || c.currentIndexInPage == len(c.pages[c.currentPage]) {
		if len(c.pages) > 0 && !c.usesPaginationCursors {
			return false
		}
		c.finishedReadingPage = true
	} else if !c.finishedReadingPage {
		return true
	}

	respBody, cursor, err := c.accessInfo.collectionRequest(method, fmt.Sprintf("%s?%s&%s", path, params, c.pagination), body)
	if err != nil {
		c.pendingError = err
		return true
	}
	err = json.Unmarshal([]byte(respBody), &container)
	if err != nil {
		c.pendingError = err
		return true
	}
	c.pagination = *cursor
	page := reflect.ValueOf(container).Elem()
	if page.Len() == 0 {
		return false
	}
	z := make([]interface{}, page.Len())
	for i := 0; i < page.Len(); i++ {
		z[i] = page.Index(i)
	}
	var y [][]interface{}
	a := reflect.ValueOf(reflect.ValueOf(c.pages).Interface())
	c.pages = reflect.Append(a, reflect.ValueOf(z)).Convert(reflect.TypeOf(y)).Interface().([][]interface{})
	c.size++
	c.currentIndexInPage = 0
	c.currentPage++
	c.finishedReadingPage = false
	return true
}

// next gets the next element from the pageableCollection.
func (c *pageableCollection) next() (reflect.Value, error) {
	if c.pendingError != nil {
		return reflect.ValueOf(nil), c.pendingError
	}
	elem := c.pages[c.currentPage][c.currentIndexInPage]
	c.currentIndexInPage++
	return elem.(reflect.Value), nil
}
