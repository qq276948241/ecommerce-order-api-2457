# 电商订单系统核心逻辑说明

> 本文档用于帮助新成员快速理解整个订单系统的架构、状态流转、数据模型和调用链。

---

## 📁 一、项目架构与目录结构

### 技术栈
- **语言**: Go 1.25+
- **Web 框架**: Gin v1.12
- **ORM**: GORM v1.31
- **数据库**: MySQL 8.0+
- **鉴权**: JWT (golang-jwt v5)
- **密码加密**: bcrypt

### 分层架构（三层）

```
┌───────────────────────────────────────────────┐
│  Handler 层 (HTTP 入口 / 请求解析 / 响应封装)  │
│  auth / product / cart / order / payment /    │
│  refund handler                                │
└─────────────────────┬─────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────┐
│  Service 层 (业务逻辑 / 事务 / 校验 / 状态机)  │
│  auth / product / cart / order / payment /    │
│  refund service                                │
└─────────────────────┬─────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────┐
│  Repository 层 (数据访问 / CRUD / 查询封装)    │
│  user / product / cart / order / payment /    │
│  refund repo                                   │
└───────────────────────────────────────────────┘
```

### 目录结构

```
project45/
├── cmd/
│   └── main.go                 # 程序入口 & 依赖注入 & 路由注册
├── internal/
│   ├── config/                 # 配置加载
│   ├── database/               # GORM 连接 & AutoMigrate
│   ├── handler/                # HTTP 处理器（6个）
│   ├── middleware/             # JWT 中间件
│   ├── model/                  # 数据模型（7个表）
│   ├── repository/             # 数据访问层（6个）
│   └── service/                # 业务逻辑层（6个）
├── pkg/
│   ├── dbutil/                 # 通用数据库操作封装（泛型）
│   ├── errors/                 # 统一业务错误（错误码+HTTP状态码）
│   └── response/               # 统一响应封装
├── sql/
│   └── init.sql                # 建表SQL + 初始数据
├── go.mod
└── understanding.md            # 本文档
```

### 依赖注入（main.go 中完成）

所有三层通过构造函数注入，无全局变量，便于单元测试：

```go
// main.go L26-L45
userRepo := repository.NewUserRepository(db)
authService := service.NewAuthService(userRepo, &cfg.JWT)
authHandler := handler.NewAuthHandler(authService)
```

---

## 🔄 二、订单状态机

### 状态枚举（order.go L5-L12）

| 状态值 | 常量名 | 说明 |
|--------|--------|------|
| 0 | `OrderStatusPending` | 待支付（下单成功后的初始状态） |
| 1 | `OrderStatusPaid` | 已支付 |
| 2 | `OrderStatusShipped` | 已发货（后台手动操作，暂未实现接口） |
| 3 | `OrderStatusCompleted` | 已完成（后台手动操作，暂未实现接口） |
| 4 | `OrderStatusCancelled` | 已取消（用户主动取消/超时未支付） |
| 5 | `OrderStatusRefunded` | 已退款（用户申请退款后） |

### 状态流转图

```
                    ┌─────────────────┐
                    │  Pending (0)    │ ◄─── 下单成功
                    └────────┬────────┘
                             │
           ┌─────────────────┼─────────────────┐
           ▼                 ▼                 ▼
   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
   │  Paid (1)    │  │ Cancelled(4) │  │（超时自动取   │
   │ 支付成功     │  │ 用户主动取消  │  │  消，暂未实现）│
   └──────┬───────┘  └──────────────┘  └──────────────┘
          │
          ▼
   ┌──────────────┐
   │ Shipped (2)  │  发货（后台）
   └──────┬───────┘
          │
          ▼
   ┌──────────────┐
   │ Completed(3) │  确认收货（后台）
   └──────┬───────┘
          │
          ▼
   ┌──────────────┐
   │ Refunded (5) │  申请退款（用户）
   └──────────────┘
```

### 关键规则

1. **Pending → Paid**：调用支付接口，支付成功后流转（见 payment_service.go）
2. **Pending → Cancelled**：调用取消订单接口，仅 Pending 状态可取消（见 order_service.go L190-L236）
3. **Paid/Completed → Refunded**：调用退款接口，申请后流转（见 refund_service.go L36-L103）
4. **所有状态流转都有行锁 + 乐观条件**：`UPDATE ... WHERE status = 原状态`，并发安全
5. **不可逆**：Paid 不能变回 Pending，Cancelled/Refunded 后不可再操作

---

## 📊 三、数据库 ER 图

### 7 张表关系

```mermaid
erDiagram
    users ||--o{ cart_items : "拥有"
    users ||--o{ orders : "下单"
    users ||--o{ payments : "支付"
    users ||--o{ refunds : "退款"

    products ||--o{ cart_items : "被加入"
    products ||--o{ order_items : "被购买"

    orders ||--|{ order_items : "包含"
    orders ||--o| payments : "有一个"
    orders ||--o| refunds : "有一个"

    users {
        bigint id PK
        varchar username UK
        varchar password
        varchar email UK
        varchar nickname
        datetime created_at
        datetime updated_at
    }

    products {
        bigint id PK
        varchar name
        text description
        decimal price
        int stock
        int stock_warning_threshold "默认10，低库存预警阈值"
        varchar image_url
        varchar category
        tinyint status "1-上架 0-下架"
        datetime created_at
        datetime updated_at
    }

    cart_items {
        bigint id PK
        bigint user_id FK
        bigint product_id FK
        int quantity
        datetime created_at
        datetime updated_at
        unique uk_user_product "(user_id, product_id)"
    }

    orders {
        bigint id PK
        varchar order_no UK
        bigint user_id FK
        decimal total_amount
        tinyint status "0-5 见状态机"
        varchar address
        varchar remark "订单备注"
        datetime pay_time
        datetime created_at
        datetime updated_at
    }

    order_items {
        bigint id PK
        bigint order_id FK
        bigint product_id FK
        varchar product_name "快照，商品后续改名不影响历史订单"
        decimal price "快照"
        int quantity
        decimal subtotal
    }

    payments {
        bigint id PK
        varchar payment_no UK
        bigint order_id FK UK "一个订单只能有一条支付记录"
        bigint user_id FK
        decimal amount
        varchar payment_method
        tinyint status "0-待支付 1-成功 2-失败"
        varchar transaction_id
        datetime paid_at
        datetime created_at
        datetime updated_at
    }

    refunds {
        bigint id PK
        varchar refund_no UK
        bigint order_id FK UK "一个订单只能申请一次退款"
        bigint user_id FK
        decimal amount
        varchar reason
        tinyint status "0-待审核 1-通过 2-拒绝 3-完成"
        varchar remark
        datetime refunded_at
        datetime created_at
        datetime updated_at
    }
```

### 设计要点

1. **订单快照**：`order_items` 存 `product_name` 和 `price` 快照，防止商品后续修改影响历史订单
2. **唯一约束**：`payments.order_id`、`refunds.order_id` 都是 UNIQUE，数据库层面防重复
3. **用户隔离**：所有业务表都有 `user_id`，接口层通过 JWT 拿用户 ID 做数据隔离
4. **低库存标记**：`products.stock_warning_threshold` 字段，`LowStock` 计算字段（不存库）

---

## 🔗 四、核心接口调用链（从下单到支付完整流程）

### 涉及接口

| 步骤 | 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|------|
| 1 | POST | `/api/auth/login` | 否 | 登录拿 JWT |
| 2 | GET | `/api/products` | 否 | 浏览商品 |
| 3 | POST | `/api/cart` | 是 | 加入购物车 |
| 4 | POST | `/api/orders` | 是 | 提交订单（核心） |
| 5 | POST | `/api/payments/pay` | 是 | 支付（核心） |
| 6 | POST | `/api/refunds` | 是 | 申请退款（可选） |

---

### 步骤 4：下单接口 POST `/api/orders` 完整调用链

```
HTTP 请求
    │
    ▼
┌──────────────────────────────────────┐
│  order_handler.CreateOrder()         │
│  - 解析 JWT 拿 user_id               │
│  - 绑定 CreateOrderRequest           │
│  - 调用 service.CreateOrder          │
└──────────────┬───────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│  order_service.CreateOrder(userID, req)                     │
│  TRANSACTION  ──────────────────────────────────────────┐   │
│  │  1. 查购物车（tx.Preload("Product")）                │   │
│  │  2. SELECT ... FOR UPDATE 锁商品行（防止并发）        │   │
│  │  3. 校验：商品状态、库存 >= 购买数量                  │   │
│  │  4. 计算总金额，生成订单项快照                        │   │
│  │  5. 生成订单号（ORD + 用户ID + 时间 + 纳秒）          │   │
│  │  6. INSERT orders                                    │   │
│  │  7. INSERT order_items                               │   │
│  │  8. UPDATE products SET stock = stock - ?            │   │
│  │        WHERE id = ? AND stock >= ?                   │   │
│  │        （乐观锁，RowsAffected=0 则回滚）              │   │
│  │  9. DELETE cart_items WHERE id IN (...)              │   │
│  └──────────────────────────────────────────────────────┘   │
│  - 错误处理：BizError 透传，其他 WrapInternal               │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│  response.Fail(c, err) / Success()   │
│  - 自动识别 BizError，返回对应 HTTP 码│
└──────────────────────────────────────┘
```

**核心代码位置**：
- Handler: [order_handler.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/handler/order_handler.go)
- Service: [order_service.go L37-L155](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/service/order_service.go#L37-L155)

---

### 步骤 5：支付接口 POST `/api/payments/pay` 完整调用链

```
HTTP 请求
    │
    ▼
┌──────────────────────────────────────┐
│  payment_handler.Pay()               │
│  - 解析 JWT 拿 user_id               │
│  - 绑定 PayRequest                   │
│  - 调用 service.Pay                  │
└──────────────┬───────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│  payment_service.Pay(userID, req)                           │
│  TRANSACTION  ──────────────────────────────────────────┐   │
│  │  1. SELECT ... FOR UPDATE 锁 orders 行               │   │
│  │  2. 校验：用户归属、status == Pending                 │   │
│  │  3. SELECT ... FOR UPDATE 锁 payments 行             │   │
│  │  4. 检查是否已有成功支付（防止重复）                  │   │
│  │  5. 创建/更新 payments 记录为 Pending                 │   │
│  │  6. mockPaymentInTx():                               │   │
│  │     ├─ payment.status = Success                      │   │
│  │     ├─ UPDATE payments SET ...                       │   │
│  │     └─ UPDATE orders SET status=Paid, pay_time=?     │   │
│  │           WHERE id=? AND status=Pending              │   │
│  │           （乐观锁，RowsAffected=0 则回滚）           │   │
│  └──────────────────────────────────────────────────────┘   │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
         响应结果
```

**核心代码位置**：
- Service: [payment_service.go L36-L137](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/service/payment_service.go#L36-L137)

---

## 🔒 五、并发控制与安全机制

### 1. 超卖防护（下单）

| 防护层级 | 技术 | 说明 |
|----------|------|------|
| 第1层 | `SELECT ... FOR UPDATE` | 事务内对商品行加排他锁，其他请求必须等事务结束 |
| 第2层 | 乐观更新条件 | `UPDATE ... WHERE stock >= 购买数量`，RowsAffected=0 回滚 |
| 第3层 | 全事务包裹 | 从查库存到扣减到创建订单在一个事务内，原子性 |

**代码位置**：[order_service.go L45-L144](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/service/order_service.go#L45-L144)

### 2. 重复支付防护

| 防护层级 | 技术 | 说明 |
|----------|------|------|
| 第1层 | `SELECT ... FOR UPDATE` | 锁订单行 + 支付记录行 |
| 第2层 | 乐观状态更新 | `UPDATE orders SET status=Paid WHERE id=? AND status=Pending` |
| 第3层 | 数据库唯一约束 | `payments.order_id` UNIQUE KEY |
| 第4层 | 业务校验 | 支付前检查已有支付记录的状态 |

### 3. 重复退款防护

同支付：`refunds.order_id` UNIQUE KEY + 行锁 + 乐观更新。

### 4. 订单状态并发修改（取消订单）

| 防护层级 | 技术 | 说明 |
|----------|------|------|
| 第1层 | `SELECT ... FOR UPDATE` | 锁订单行 |
| 第2层 | 乐观状态更新 | `UPDATE orders SET status=Cancelled WHERE id=? AND status=Pending` |
| 第3层 | 库存回滚 | 事务内回补 `stock = stock + quantity`，原子性 |

**代码位置**：[order_service.go L190-L236](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/service/order_service.go#L190-L236)

### 5. 用户数据隔离

- JWT 中间件解析 token 拿到 `user_id`，存到 Gin Context
- 所有业务接口从 Context 拿 `user_id`，SQL 条件必带 `user_id = ?`
- 权限校验：`if order.UserID != userID { return 403 }`

**代码位置**：[jwt.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/middleware/jwt.go)

---

## 🎯 六、公共组件说明

### 1. pkg/errors — 统一业务错误

```go
type BizError struct {
    Code     int    // 业务错误码 40000/40100/40300/40400/40900/50000
    Message  string // 业务信息
    HTTPCode int    // HTTP 状态码
    Err      error  // 内部原始错误（打日志用）
}

// 构造函数
bizerr.BadRequest("参数错误")
bizerr.BadRequestf("商品 %s 库存不足", name)
bizerr.NotFound("订单不存在")
bizerr.Forbidden("无权操作")
bizerr.Conflict("用户名已存在")
bizerr.WrapInternal(err, "创建订单失败")
```

**代码位置**：[errors.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/pkg/errors/errors.go)

### 2. pkg/dbutil — 通用数据库封装（泛型）

```go
// 条件构造
dbutil.Eq("user_id", 1)
dbutil.In("id", ids)
dbutil.Gt("stock", 0)
dbutil.Like("name", "iPhone")

// 通用 CRUD
dbutil.Create(db, &entity)
dbutil.GetByID[model.Order](db, id, "Items")    // 支持 Preload
dbutil.GetOne[model.User](db, []Condition{...})
dbutil.List[model.Product](db, conds, orderBys, preloads...)
dbutil.Paginate[model.Order](db, pageQuery, conds, orderBys)
dbutil.Update(db, &entity)
dbutil.UpdateFields(db, model, conds, fields)    // 条件更新 map
dbutil.DeleteByID(db, &model.Product{}, id)
dbutil.Count / Exists
```

**代码位置**：[dbutil.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/pkg/dbutil/dbutil.go)

### 3. pkg/response — 统一响应

```go
// 成功
response.Success(c, data)            // 200 {code:0, message:"success", data:...}

// 错误（自动识别 BizError）
response.Fail(c, err)                // 自动返回对应 HTTP 码和业务码
```

**代码位置**：[response.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/pkg/response/response.go)

---

## 📌 七、快速定位表

| 功能 | 模型文件 | 状态常量位置 |
|------|----------|--------------|
| 用户 | [user.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/model/user.go) | - |
| 商品（含库存预警） | [product.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/model/product.go) | `DefaultStockWarningThreshold = 10` |
| 购物车 | [cart.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/model/cart.go) | - |
| 订单 | [order.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/model/order.go) | `OrderStatusPending` 等 L5-L12 |
| 支付 | [payment.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/model/payment.go) | `PaymentStatusPending` 等 L5-L9 |
| 退款 | [refund.go](file:///d:/code/ai-prompt/solo-20/repos/repo45/project45/internal/model/refund.go) | `RefundStatusPending` 等 L5-L10 |

---

## 🚀 八、启动说明

```bash
# 1. 建库
mysql < sql/init.sql

# 2. 配置环境变量（可选，有默认值）
export DB_HOST=127.0.0.1
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=123456
export DB_NAME=ecommerce
export JWT_SECRET=your-secret
export SERVER_PORT=8080

# 3. 运行
go run cmd/main.go
```

启动后访问 `http://localhost:8080`，所有 API 都在 `/api` 路径下。
