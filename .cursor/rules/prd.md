### **1. Triết lý thiết kế (Design Philosophy)**

Trước khi đi vào code, chúng ta cần xác định các nguyên tắc cốt lõi:

- **Decoupling (Tách rời):** Phần điều phối (dispatcher) và thực thi (worker) không cần biết chi tiết logic của job. Ngược lại, job cũng không cần biết nó đang được chạy bởi worker nào. Điều này đạt được thông qua **Interfaces**.
- **Type Safety:** Sử dụng **Go Generics** (từ Go 1.18+) để đảm bảo an toàn kiểu dữ liệu cho "payload" của mỗi job. Tránh việc phải ép kiểu `interface{}` một cách không an toàn.
- **Concurrency Control:** Người dùng phải có khả năng kiểm soát số lượng goroutine chạy đồng thời để tránh làm quá tải hệ thống. Đây là lúc **Worker Pool Pattern** phát huy tác dụng.
- **Graceful Shutdown:** Hệ thống phải có khả năng tắt một cách "duyên dáng", đảm bảo các job đang chạy được hoàn thành trước khi chương trình thoát. `context.Context` là công cụ hoàn hảo cho việc này.
- **Configurable & Extensible (Dễ cấu hình và mở rộng):** Sử dụng **Functional Options Pattern** để việc cấu hình dispatcher trở nên linh hoạt và dễ đọc.

---

### **2. Các thành phần chính (Core Components)**

Package của chúng ta sẽ xoay quanh 3 thành phần chính.

#### **a. `Job[T any]` Interface**

Đây là trái tim của sự trừu tượng. Bất kỳ tác vụ nào muốn được xử lý bởi hệ thống của chúng ta đều phải implement interface này. Việc sử dụng generic `[T any]` cho phép chúng ta định nghĩa một kiểu dữ liệu cụ thể (payload) cho mỗi loại job.

```go
package asyncjob

import "context"

// Job là một interface định nghĩa một đơn vị công việc có thể thực thi.
// T là kiểu dữ liệu của payload mà job sẽ xử lý.
type Job[T any] interface {
    // Execute chứa logic thực thi của job.
    // Nó nhận vào một context để xử lý timeout hoặc cancellation,
    // và một payload với kiểu dữ liệu T.
    Execute(ctx context.Context, payload T) error
}
```

- **Tại sao lại dùng `Job[T]`?** Nó giúp logic `Execute` nhận được đúng kiểu dữ liệu `payload` mà không cần type assertion, giúp code an toàn và dễ maintain hơn.

#### **b. `Dispatcher[T any]` (Worker Pool)**

Đây là thành phần điều phối, chịu trách nhiệm quản lý một "pool" các `Worker` và phân phát `Job` cho chúng.

```go
package asyncjob

import "context"

// Dispatcher quản lý một pool các worker để xử lý job một cách bất đồng bộ.
// T là kiểu dữ liệu của payload mà dispatcher này sẽ xử lý.
type Dispatcher[T any] struct {
    // Kênh để nhận payload công việc.
    jobPayloadQueue chan T

    // Kênh bên trong để worker nhận job đã được đóng gói.
    internalJobQueue chan jobHolder[T]

    // Job handler chứa logic thực thi.
    jobHandler Job[T]

    // Context để quản lý vòng đời của dispatcher.
    ctx    context.Context
    cancel context.CancelFunc

    // WaitGroup để chờ tất cả các goroutine kết thúc.
    wg sync.WaitGroup
}

// jobHolder là một struct nội bộ để đóng gói payload và context.
type jobHolder[T any] struct {
    payload T
    ctx     context.Context
}
```

- **`jobPayloadQueue`:** Một kênh (channel) mà bên ngoài sẽ gửi `payload` vào.
- **`internalJobQueue`:** Kênh mà các `Worker` sẽ lắng nghe để lấy việc.
- **`jobHandler`:** Một thể hiện của `Job[T]` sẽ được tất cả các worker sử dụng để `Execute`. Điều này đảm bảo một dispatcher chỉ xử lý một loại logic job, giúp thiết kế gọn gàng.
- **`ctx`, `cancel`, `wg`:** Các công cụ chuẩn của Go để quản lý concurrency và graceful shutdown.

#### **c. `Worker[T any]`**

Worker là một goroutine thực thi công việc. Nó không phải là một struct public mà là một phần bên trong của `Dispatcher`.

Logic của một worker sẽ như sau (dưới dạng mã giả):

```go
// function run của worker
func (d *Dispatcher[T]) worker() {
    defer d.wg.Done()
    for {
        select {
        case job := <-d.internalJobQueue:
            // Thực thi job với payload và context của nó
            if err := d.jobHandler.Execute(job.ctx, job.payload); err != nil {
                // Xử lý lỗi ở đây: log, gửi vào một kênh lỗi, etc.
                // Tùy chọn này có thể được cấu hình qua Options.
            }
        case <-d.ctx.Done():
            // Nhận tín hiệu shutdown, thoát khỏi vòng lặp.
            return
        }
    }
}
```

---

### **3. Triển khai với Clean Code & Design Patterns**

#### **a. Functional Options Pattern**

Để khởi tạo `Dispatcher` một cách linh hoạt, chúng ta sẽ dùng Functional Options.

```go
package asyncjob

// Option là một hàm để cấu hình cho Dispatcher.
type Option[T any] func(*Dispatcher[T])

// WithMaxWorkers đặt số lượng worker sẽ chạy đồng thời.
func WithMaxWorkers[T any](count int) Option[T] {
    return func(d *Dispatcher[T]) {
        if count > 0 {
            // Logic để set số worker (sẽ được dùng trong NewDispatcher)
        }
    }
}

// WithQueueSize đặt kích thước buffer của hàng đợi job.
func WithQueueSize[T any](size int) Option[T] {
    return func(d *Dispatcher[T]) {
        if size > 0 {
            d.jobPayloadQueue = make(chan T, size)
        }
    }
}

// WithErrorHandler cung cấp một hàm callback để xử lý lỗi từ job.
func WithErrorHandler[T any](handler func(payload T, err error)) Option[T] {
    return func(d *Dispatcher[T]) {
        // Logic để gán error handler
    }
}
```

#### **b. Constructor `NewDispatcher`**

Hàm khởi tạo sẽ gom tất cả lại với nhau.

```go
package asyncjob

import (
    "context"
    "sync"
)

// NewDispatcher tạo và khởi chạy một dispatcher mới.
func NewDispatcher[T any](jobHandler Job[T], opts ...Option[T]) *Dispatcher[T] {
    // Giá trị mặc định
    const defaultMaxWorkers = 10
    const defaultQueueSize = 100

    ctx, cancel := context.WithCancel(context.Background())

    d := &Dispatcher[T]{
        jobHandler:      jobHandler,
        jobPayloadQueue: make(chan T, defaultQueueSize),
        ctx:             ctx,
        cancel:          cancel,
        // Các trường khác...
    }

    // Áp dụng các cấu hình tùy chỉnh
    for _, opt := range opts {
        opt(d)
    }

    // Khởi chạy các worker
    maxWorkers := /* lấy từ config, ví dụ: */ defaultMaxWorkers
    d.internalJobQueue = make(chan jobHolder[T])

    d.wg.Add(maxWorkers)
    for i := 0; i < maxWorkers; i++ {
        go d.worker()
    }

    // Khởi chạy goroutine trung chuyển
    d.wg.Add(1)
    go d.runDistributor()

    return d
}

// Các method của Dispatcher
func (d *Dispatcher[T]) Submit(payload T) {
    d.jobPayloadQueue <- payload
}

func (d *Dispatcher[T]) Shutdown() {
    close(d.jobPayloadQueue) // Dừng nhận thêm job mới
    d.wg.Wait()              // Chờ tất cả goroutine kết thúc
}

// ... worker và các hàm nội bộ khác
```

---

### **4. Cách tái sử dụng (Usage Example)**

Giả sử chúng ta có một dự án cần gửi email thông báo một cách bất đồng bộ.

**Bước 1: Định nghĩa Payload và Job**

```go
// In project A (e.g., in package `notifications`)

// 1. Định nghĩa Payload
type EmailPayload struct {
    To      string
    Subject string
    Body    string
}

// 2. Implement interface asyncjob.Job
type EmailSenderJob struct {
    // Có thể inject các dependency ở đây, ví dụ một email service client
    // mailClient *some_mail.Client
}

func (j *EmailSenderJob) Execute(ctx context.Context, payload EmailPayload) error {
    fmt.Printf("Sending email to %s: Subject: %s\n", payload.To, payload.Subject)

    // Logic gửi email thực tế
    // time.Sleep(1 * time.Second) // Giả lập công việc tốn thời gian

    // Trả về lỗi nếu có
    // if some_error {
    //   return errors.New("failed to send email")
    // }

    fmt.Printf("Email to %s sent successfully!\n", payload.To)
    return nil
}
```

**Bước 2: Sử dụng `Dispatcher` trong ứng dụng**

```go
package main

import (
    "fmt"
    "time"
    "your-go-module/asyncjob" // Import package bạn vừa tạo
    "your-go-module/notifications"
)

func main() {
    fmt.Println("Starting application...")

    // 1. Khởi tạo job handler
    emailJob := &notifications.EmailSenderJob{}

    // 2. Khởi tạo dispatcher với kiểu EmailPayload
    // Sử dụng functional options để cấu hình
    dispatcher := asyncjob.NewDispatcher[notifications.EmailPayload](
        emailJob,
        asyncjob.WithMaxWorkers[notifications.EmailPayload](5),     // 5 worker đồng thời
        asyncjob.WithQueueSize[notifications.EmailPayload](1000), // Hàng đợi chứa 1000 job
    )

    fmt.Println("Dispatcher started. Submitting jobs...")

    // 3. Gửi các job vào hệ thống
    for i := 0; i < 20; i++ {
        payload := notifications.EmailPayload{
            To:      fmt.Sprintf("user%d@example.com", i),
            Subject: "Welcome!",
            Body:    "Hello there!",
        }
        dispatcher.Submit(payload)
    }

    fmt.Println("All jobs submitted. Waiting for processing...")

    // Cho các job có thời gian xử lý
    time.Sleep(3 * time.Second)

    // 4. Shutdown hệ thống một cách an toàn
    fmt.Println("Shutting down dispatcher...")
    dispatcher.Shutdown()
    fmt.Println("Application finished.")
}
```

---

### **5. Các cải tiến nâng cao (Advanced Considerations)**

Với vai trò Senior, chúng ta cần nghĩ xa hơn:

1.  **Job có kết quả trả về (Jobs with Results):** Thiết kế hiện tại theo mô hình "fire-and-forget". Nếu cần kết quả, ta có thể mở rộng `Submit` để trả về một `Future[R]` (với `R` là kiểu dữ liệu trả về), mà bên trong nó chứa một channel để nhận kết quả sau khi job hoàn thành.
2.  **Retry và Backoff:** Thêm một `Option` để cấu hình chính sách retry (ví dụ: `WithRetryPolicy(maxRetries int, backoffStrategy Backoff)`) khi `Execute` trả về lỗi.
3.  **Metrics & Observability:** Thêm các `Option` để inject một interface logger hoặc một Prometheus registry. Bên trong `Dispatcher`, ta có thể theo dõi các metrics như: số lượng job trong hàng đợi (`gauge`), tổng số job đã xử lý (`counter`), số job lỗi (`counter`), thời gian xử lý job (`histogram`).
4.  **Job có độ ưu tiên (Priority Jobs):** Thay thế channel `internalJobQueue` bằng một cấu trúc dữ liệu hàng đợi ưu tiên (priority queue).
5.  **Lưu trữ và phục hồi (Persistence):** Để đảm bảo không mất job khi ứng dụng khởi động lại, có thể thêm một tầng lưu trữ (ví dụ: Redis, aof) để lưu trạng thái của hàng đợi job.

### **Kết luận**

Bằng cách tiếp cận này, chúng ta đã tạo ra một package `asyncjob` mạnh mẽ:

- **Generic & Type-Safe:** An toàn và linh hoạt nhờ Generics.
- **Reusable:** Cực kỳ dễ dàng để tích hợp vào bất kỳ dự án nào chỉ bằng cách implement interface `Job[T]`.
- **Clean & Idiomatic:** Sử dụng các pattern phổ biến và hiệu quả trong Go như Worker Pool, Functional Options và `context`.
- **Robust:** Quản lý concurrency và có cơ chế shutdown an toàn.

Đây là một nền tảng vững chắc mà từ đó bạn có thể xây dựng các tính năng phức tạp hơn tùy theo yêu cầu của từng dự án cụ thể.
