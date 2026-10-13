# linkcard

Kite 的超链接卡片插件，从 Halo 的
[plugin-editor-hyperlink-card](https://github.com/halo-sigs/plugin-editor-hyperlink-card)
迁移而来。在文章里用一个 `card` 代码块放一张链接卡片，构建时由插件的
WebAssembly 模块画成 HTML，读者浏览器里再补上缺的图标和文字。

- **三种尺寸**，对应 Halo 的三种风格：`small`（小）、`regular`（常规）、`grid`（大图）
- **构建期成卡**：卡片是构建产物里的静态 HTML，不闪、不等脚本、对搜索引擎可见
- **自动图标**：没写 `icon` 的卡片自动用 favicon 服务的网站图标
- **自动抓取**（可在设置里关闭）：没写的标题、描述、配图在浏览器端向
  Microlink 或自定义接口询问，结果缓存在读者浏览器里，一张卡片只问一次
- 插件卸载后，`card` 代码块退回为普通代码块，**不会让构建失败**，也不会泄露内容
- **裸链接也能成卡**（设置里开）：正文里链接文字就是网址本身的（如直接粘贴的
  `https://example.com`）也渲染成卡片；只作用于正文（`<article>` 内），菜单、
  页脚、带名字的链接不动。在站点 `kite.yaml` 里开：

```yaml
plugins:
  enabled: [linkcard]
  settings:
    linkcard:
      links: true
```

## 安装

从 [Releases](https://github.com/bailangvvkg/plugin-linkcard/releases) 下载最新 zip（版本号是打包的年月日时分秒；里面有编译好的 `plugin.wasm`，不需要 Go），在站点目录：

```bash
kite plugin add <下载的zip路径>    # 装进 plugins/linkcard，默认关着
kite plugin enable linkcard       # 打开，写入 kite.yaml
```

或者从本仓库的源码安装（需要 Go 1.24+）：

```bash
git clone https://github.com/bailangvvkg/plugin-linkcard
cd plugin-linkcard
GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -trimpath -ldflags '-s -w' -o plugin.wasm .
cd 你的站点
kite plugin add ../plugin-linkcard
kite plugin enable linkcard
```

然后在站点目录检查：

```bash
kite plugin verify plugins/linkcard
```

在后台“应用中心”（或插件设置页）确认插件已启用，按需调整设置。

## 用法

在文章里写一个 `card` 代码块，一行一项，`键: 值` 或 `键=值` 都行：

````markdown
```card
url: https://status.anye.xyz/status/docker
title: Docker 状态页
desc: 各服务的运行状态与可用性监控
size: grid
```
````

最省事的写法是只给地址，其余交给浏览器端补全：

````markdown
```card
https://www.kite.plus
```
````

| 参数 | 别名 | 说明 |
|---|---|---|
| `url` | `link`、`href` | 链接地址，http(s) 绝对地址或站内 `/` 开头的路径；也可以不写键名，直接把地址写在第一行 |
| `title` | `name` | 标题；不写时先用网址主机名顶着，浏览器端抓到后替换 |
| `desc` | `description`、`summary` | 描述；`small` 尺寸不显示 |
| `image` | `img`、`cover` | 配图；`regular` 显示在卡片右侧，`grid` 显示在顶部 |
| `icon` | `favicon`、`logo` | 图标；不写时自动用 favicon 服务 |
| `size` | `theme` | `small` / `regular` / `grid`，默认用设置里的“默认尺寸” |
| `target` |  | `_blank` / `_self`，默认用设置里的“链接打开方式” |

参数值可以用引号包起来；`#` 开头的行是注释。多个卡片挨着写就是一组。

### 三种尺寸

- **small** —— 一行小条：图标 + 标题 + 域名，适合塞在段落之间
- **regular** —— Halo 常规卡片：左边图标、标题、描述，右边配图（有则显示）
- **grid** —— 大图卡片：顶部 16:9 配图，下方图标、标题、描述

## 设置

后台插件设置里可以调：默认尺寸、链接打开方式、是否浏览器端抓取元数据、
抓取来源（Microlink 或自定义接口 `{url}` 占位符）、是否缓存抓取结果、图标服务
（DuckDuckGo / Google / 不使用）。

Microlink 免费额度按来源 IP 计，每天约 50 次；开启缓存后读者每张卡只问一次。
也可以自建或换用任何返回 `title` / `description` / `image` / `logo` 字段 JSON
的接口，在设置里选“自定义接口”并填地址。

## 从 Halo 迁移的文章

从 Halo 导入的文章里已经写下的 `<hyperlink-card>` 和 `<hyperlink-inline-card>`
标签（连同 `custom-title` / `custom-description` / `custom-image` / `custom-icon`
属性，`theme` 对应 `size`，inline 卡片算 small）也会被本插件转成同样的卡片。

有一点要注意：这些标签是正文里的原始 HTML，需要站点的
`markdown: {unsafeHTML: true}` 打开才会进到页面；上面的 `card` 代码块
则不需要打开它。

## 和 Halo 版的差异

| Halo | 本插件 |
|---|---|
| 编辑器里选中链接转换卡片 | 写 `card` 代码块（Kite 插件目前不能扩展编辑器界面） |
| 服务端抓取链接详情，缓存 12 小时 | 构建期不抓取（WASM 沙箱无网络）；缺什么在浏览器端补 |
| `<hyperlink-card>` Web Component | 静态 HTML，脚本只负责补缺 |
| 编辑器里选尺寸 | `size` 参数或后台“默认尺寸”设置 |

## 许可

MIT
