package utils

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// 优化后的ID生成器
var (
	snowflakeInstance *SnowflakeIdWorker
	once              sync.Once
	mutex             sync.Mutex
	protoNumberCache  sync.Map
)

// int32转换成字节
func IntToBytes(n int32) []byte {
	x := n
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// long转换成字节
func LongToBytes(n int64) []byte {
	x := n
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// short转换成字节
func ShortToBytes(n int16) []byte {
	x := n
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// 字节转换成int
func BytesToInt(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

// 字节转换成long
func BytesToLong(b []byte) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int64
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

// 字节转换成long
func BytesToShort(b []byte) int16 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int16
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

// 把协议名称转为唯一协议编号
func ProtocalNumber(replacement string) int32 {
	// 先从缓存获取
	if val, ok := protoNumberCache.Load(replacement); ok {
		return val.(int32)
	}

	// 计算哈希值
	var h int32
	for _, char := range replacement {
		h = 31*h + int32(char)
	}

	// 缓存结果
	protoNumberCache.Store(replacement, h)
	return h
}

// 获取协议名称
func GetProtoName(t proto.Message) string {
	x := proto.MessageName(t)
	proto_ := strings.Split(string(x), ".")
	if len(proto_) > 0 {
		return proto_[1]
	}
	return ""
}

// 连个字符串的key合并
func MergeMapString(varA map[string]string, varB map[string]string) map[string]string {
	data := make(map[string]string, len(varA)+len(varB))
	for k, v := range varA {
		data[k] = v
	}
	for k, v := range varB {
		data[k] = v
	}
	return data
}

// 读取csv数据
func ReadCSVData(_file string) [][]string {
	mutex.Lock()
	csvFile, err := os.Open(_file)
	defer csvFile.Close()
	if nil != err {
		return nil
	}
	data := csv.NewReader(bufio.NewReader(csvFile))
	var LineNumber int
	var dataRecords [][]string
	for {
		record, err := data.Read()
		// 如果读到文件的结尾，EOF的优先级居然比nil还高！
		if err == io.EOF {
			break
		} else if err != nil {
			return nil
		}
		// Read返回的是一个数组，它已经帮我们分割了，
		if LineNumber > 2 {
			s := make([]string, len(record))
			for i := 0; i < len(record); i++ {
				s[i] = record[i]
			}
			dataRecords = append(dataRecords, s)
		}
		LineNumber++
	}
	mutex.Unlock()
	return dataRecords
}

// 读取txt文件
func ReadTXTData(_file string) []string {
	mutex.Lock()
	file, err := os.Open(_file)
	if nil != err {
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var dataRecords []string
	//是否有下一行
	for scanner.Scan() {
		dataRecords = append(dataRecords, scanner.Text())
	}
	mutex.Unlock()
	return dataRecords
}

func GetGid() int64 {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Errorf("can not get goroutine id: %v", err))
	}
	return int64(id)
}

func GenSnowflakeToo() uint64 {
	once.Do(initSnowflake)
	if snowflakeInstance == nil {
		return 0
	}
	return snowflakeInstance.NextId()
}
