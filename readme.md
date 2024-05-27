<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/auto-withdraw-advideo">
    <img src="img/1.jpg" width="200" height="200" alt="auto-withdraw-advideo">
  </a>
</p>

<div align="center">
# QQ群二维码视频广告过滤器

_✨ 适用于Onebotv11的一键端 ✨_  
</div> 

---

## 介绍

最近实在是受够了,那些QQ群半夜混进的小号,发送的只有几秒钟的视频二维码广告。

把群环境弄得一团糟,而且基本都是出现在半夜3点左右,QQ的Q群管家也无法识别1-3秒的视频情色二维码广告.

本项目旨在提供一个简单而有效的解决方案，自动识别并撤回群内的短视频广告，确保群聊的清洁和用户的不被打扰。

目前采用判断视频长度的方案,可能会有误伤,

本工具能够识别并判断配置中指定秒数以内的视频（时间长度可自定义），让半夜的视频广告无处遁形，即便他们混进了群里也无济于事。

## 主要功能
- **自动撤回短视频广告**：自动检测并撤回指定秒数以内的视频。
- **配置极简**：用户只需要简单配置即可开始使用。
- **支持Onebot v11标凈**：适配使用Onebot v11标准的机器人。

## 用法
要使用本工具，你需要进行简单的配置：

- `port`: 例如 `"28800"` 表示监听 `ws://127.0.0.1:28800`。
- `video_second_limit`: 例如 `5` 代表自动撤回5秒以内的视频。

确保机器人连接上这个反向WebSocket地址，并设置机器人为群管理员，即可开始自动工作。

### 示例配置
```yaml
port: "28800"
video_second_limit: 5
```

## 使用方法
1. 确保机器人使用的是支持Onebot v11的实现。
2. 配置机器人连接到指定的WebSocket地址。
3. 将机器人设置为群管理员。
4. 调整`video_second_limit`配置以撤回指定长度的视频广告。

## TODO
- 拦截并撤回更多类型的广告。
- 实现进群验证码功能。
- 自定义撤回规则。
- 撤回卡片信息等。
- [x] 找到视频广告更多特征,更精准的识别视频广告.

## 为了更准确的识别的视频二维码广告,你需要安装ffmpeg并设置环境变量

若不安装,请将check_video_qrcode改为false,但不会对视频内的二维码进行识别,精准度会下降.

## 贡献
欢迎对本项目提出改进建议或直接贡献代码，一起打造更清洁的聊天环境。

## 以下是在Windows和Linux上安装FFmpeg的详细步骤：

### Windows

#### 1. 下载FFmpeg
1. 访问FFmpeg的官方下载页面：[FFmpeg Official Download](https://ffmpeg.org/download.html)。
2. 选择适合你的Windows系统的版本。通常，你可以下载“Windows builds from gyan.dev”或者“BtbN/FFmpeg-Builds”中的版本。
3. 下载一个静态构建（static build），这样不需要额外安装任何依赖。

#### 2. 安装FFmpeg
1. 将下载的ZIP文件解压到你希望安装FFmpeg的目录，如 `C:\Program Files\FFmpeg`。

#### 3. 设置环境变量
1. 右击“此电脑”或“计算机”，选择“属性”。
2. 点击“高级系统设置”链接。
3. 在系统属性窗口中，点击“环境变量”按钮。
4. 在“系统变量”区域中找到“Path”变量，然后点击“编辑”。
5. 点击“新建”，添加FFmpeg的bin目录的路径，例如 `C:\Program Files\FFmpeg\bin`。
6. 点击“确定”保存更改。

#### 4. 验证安装
1. 打开命令提示符（CMD）。
2. 输入 `ffmpeg -version`，按回车。如果系统显示FFmpeg的版本信息，则说明安装成功。

### Linux

Linux上的安装过程取决于你使用的Linux发行版。以下是在Ubuntu和CentOS上安装FFmpeg的步骤。

#### Ubuntu

1. **更新包列表**：
   ```bash
   sudo apt update
   ```

2. **安装FFmpeg**：
   ```bash
   sudo apt install ffmpeg
   ```

3. **验证安装**：
   ```bash
   ffmpeg -version
   ```

#### CentOS

对于CentOS，你可能需要启用EPEL仓库来安装FFmpeg：

1. **启用EPEL仓库**：
   ```bash
   sudo yum install epel-release
   ```

2. **安装Nux Dextop，一个提供多媒体和桌面包的仓库**（对于老版本的CentOS 7）：
   ```bash
   sudo rpm --import http://li.nux.ro/download/nux/RPM-GPG-KEY-nux.ro
   sudo rpm -Uvh http://li.nux.ro/download/nux/dextop/el7/x86_64/nux-dextop-release-0-1.el7.nux.noarch.rpm
   ```

3. **安装FFmpeg**：
   ```bash
   sudo yum install ffmpeg ffmpeg-devel
   ```

4. **验证安装**：
   ```bash
   ffmpeg -version
   ```

这些步骤将帮助你在各自的操作系统上安装和配置FFmpeg。确保在安装过程中根据你的系统版本和需求调整命令。