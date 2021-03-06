# string

```golang
// StringHeader is the runtime representation of a string.
// It cannot be used safely or portably and its representation may
// change in a later release.
// Moreover, the Data field is not sufficient to guarantee the data
// it references will not be garbage collected, so programs must keep
// a separate, correctly typed pointer to the underlying data.
type StringHeader struct {
	Data uintptr
	Len  int
}
```

与 [slice](slice.md) 类似，它包含一个指向数组的指针`Data`和一个长度`len`来表示当前字符串的长度

![](https://raw.githubusercontent.com/codeflysafe/gitalk/main/img/20220412163221.png)

### 查看底层结构
```go
func pointer() {
	var str string
	str = "123123"
	p := (*reflect.StringHeader)(unsafe.Pointer(&str))
	fmt.Println(p)

	// 访问第一个元素
	ptr := unsafe.Pointer(p.Data)
	fmt.Println(*(*byte)(ptr), '1')
    // 转化为底层数组
	arr := (*[5]byte)(ptr)
	fmt.Println(arr)

	// 访问数组的第i个元素 地址寻址 按照字
	fmt.Println(arr[3], *(*byte)(add(ptr, 3)))
}
```

执行结果如下
```shell
&{17614562 6}
49 49
&[49 50 51 49 50]
49 49
```

### 不可变
string 是不可变的，所有对string自身的修改都是新创建了一个字符串

![](https://raw.githubusercontent.com/codeflysafe/gitalk/main/img/20220412165506.png)

```go
func change() {
	var str, str2, str3, str4 string
	str = "123123"
	p := (*reflect.StringHeader)(unsafe.Pointer(&str))

	str2 = str[0:3]
	p2 := (*reflect.StringHeader)(unsafe.Pointer(&str2))

	str3 = str[1:4]
	p3 := (*reflect.StringHeader)(unsafe.Pointer(&str3))

	// 对比 p1 和 p2 的位置, 位置是相同的, 但是p3的指针向前移动一个字长
	// 说明，与切片一样，字符串切片底层也是共享的一个底层数组
	fmt.Println(p.Data, p2.Data, p3.Data)

	// 对比数组内容, 指针没变，但是长度修改了
	arr1 := (*[5]byte)(unsafe.Pointer(p.Data))
	arr2 := (*[3]byte)(unsafe.Pointer(p2.Data))
	arr3 := (*[3]byte)(unsafe.Pointer(p3.Data))
	fmt.Println(arr1, str, p.Len)
	fmt.Println(arr2, str2, p2.Len)
	fmt.Println(arr3, str3, p3.Len)
	// 修改呢？
	// str[0] = 12 违法的操作
	//1。 采用byte[] 数组修改，然后在变成 string
	fmt.Println(" 采用byte[] 数组修改，然后在变成 string")
	c := []byte(str)
	fmt.Println(c)
	for i, _ := range c {
		// 修改，可以吗？
		c[i] = 'k'
	}
	fmt.Println(c)
	// str4 := string(c) 变回字符串
	fmt.Println(" 采用byte[] 数组修改，------- string")

	// 2. + 号的作用
	fmt.Println("采用 + 修改字符串  ------- ")
	fmt.Println(str4)
	str4 = str2 + str3
	p4 := (*reflect.StringHeader)(unsafe.Pointer(&str4))
	fmt.Println(str4, p4.Data)

	fmt.Println("采用 + 修改字符串  ------- end ")
	// 2。 stringbuild ?
	build := strings.Builder{}
	// 3
	build.WriteString(str2)
	// 3
	build.WriteString(str4)
	// buf := make([]byte, len(b.buf), 2*cap(b.buf)+n) 扩容
	fmt.Println(build.String(), build.Cap(), build.Len())
}
```

结果如下

```shell
17618242 17618242 17618243
&[49 50 51 49 50] 123123 6
&[49 50 51] 123 3
&[50 51 49] 231 3
 采用byte[] 数组修改，然后在变成 string
[49 50 51 49 50 51]
[107 107 107 107 107 107]
 采用byte[] 数组修改，------- string
采用 + 修改字符串  ------- 

123231 824633827568
采用 + 修改字符串  ------- end 
123123231 16 9



```