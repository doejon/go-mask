# go-mask

Go-mask package allows you to mask values.

The package is heavily built on top of github.com/barkimedes/go-deepcopy:
in order to mask fields, go-mask will create a deep copy of your passed value before masking.
This allows you to reuse your original object.

## Usage

You have two possibilities to mask your data:

- In case your data structure will be referenced as a pointer, you can implement the Masker interface
- in case of primivite values (e.g. strings, integers), implement a method called `func(t T) MaskXXX() T` returning the primitive type

**IMPORTANT:** In case you're implementing the Masker interface,
you need to use a pointer receiver _and_ you need to make sure your data type is passed as a pointer.

```go
// PrimitiveType implements a method MaskXXX returning PrimitiveType
type PrimitiveType string
func(p PrimitiveType) MaskXXX() PrimitiveType{
  return PrimitiveType("MASKED")
}

type SensitiveData struct{
  Primitive PrimitiveType
  Name string
}

// MaskXXX implements the mask.Masker interface
// which will modify the keys on the copy itself.
// NOTE: MaskXXX will only work in case SensitiveData is passed as a pointer
func (s *SensitiveData) MaskXXX(){
  s.Name = "MASKED"
}


func main(){
  sensitiveData := &SensitiveData{
    Primitive: "primitive",
    Name: "name",
  }
  masked := mask.Must(sensitiveData)
  log.Printf("%v", masked)
}

```
