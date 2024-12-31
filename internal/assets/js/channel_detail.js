class ChannelDetail {
    constructor() {
        this.player = null;
        this.hls = null;
        this.currentStreamUrl = null;
        this.channelName = window.CHANNEL_NAME;
        this.channelUrl = window.CHANNEL_URL;
        
        // 初始化播放器
        this.initPlayer();
        // 加载频道信息
        this.loadChannelInfo();
    }

    initPlayer() {
        const video = document.getElementById('player');
        
        // 初始化Plyr播放器，但不显示其控制栏
        this.player = new Plyr(video, {
            controls: [],
            clickToPlay: false,
            keyboard: { focused: true, global: true }
        });

        // 确保浏览器支持HLS
        if (!Hls.isSupported()) {
            console.warn('浏览器不支持HLS');
        }

        // 初始化自定义控制栏
        this.initCustomControls();
    }

    initCustomControls() {
        const playPauseBtn = document.getElementById('playPauseBtn');
        const muteBtn = document.getElementById('muteBtn');
        const volumeSlider = document.getElementById('volumeSlider');
        const fullscreenBtn = document.getElementById('fullscreenBtn');
        const progressBar = document.getElementById('progressBar');
        const progress = progressBar.parentElement;

        // 播放/暂停按钮
        playPauseBtn.addEventListener('click', () => {
            if (this.player.playing) {
                this.player.pause();
                playPauseBtn.innerHTML = '<i class="bi bi-play-fill"></i>';
            } else {
                this.player.play();
                playPauseBtn.innerHTML = '<i class="bi bi-pause-fill"></i>';
            }
        });

        // 音量控制
        volumeSlider.addEventListener('input', (e) => {
            const volume = e.target.value / 100;
            this.player.volume = volume;
            this.updateVolumeIcon(volume);
        });

        muteBtn.addEventListener('click', () => {
            this.player.muted = !this.player.muted;
            if (this.player.muted) {
                muteBtn.innerHTML = '<i class="bi bi-volume-mute"></i>';
                volumeSlider.value = 0;
            } else {
                this.updateVolumeIcon(this.player.volume);
                volumeSlider.value = this.player.volume * 100;
            }
        });

        // 全屏按钮
        fullscreenBtn.addEventListener('click', () => {
            this.player.fullscreen.toggle();
        });

        // 进度条
        progress.addEventListener('click', (e) => {
            const rect = progress.getBoundingClientRect();
            const pos = (e.clientX - rect.left) / rect.width;
            this.player.currentTime = pos * this.player.duration;
        });

        // 监听播放器事件
        this.player.on('timeupdate', () => {
            const progress = (this.player.currentTime / this.player.duration) * 100;
            progressBar.style.width = `${progress}%`;
        });

        this.player.on('play', () => {
            playPauseBtn.innerHTML = '<i class="bi bi-pause-fill"></i>';
        });

        this.player.on('pause', () => {
            playPauseBtn.innerHTML = '<i class="bi bi-play-fill"></i>';
        });

        this.player.on('enterfullscreen', () => {
            fullscreenBtn.innerHTML = '<i class="bi bi-fullscreen-exit"></i>';
        });

        this.player.on('exitfullscreen', () => {
            fullscreenBtn.innerHTML = '<i class="bi bi-fullscreen"></i>';
        });

        // 初始化音量
        this.player.volume = volumeSlider.value / 100;
        this.updateVolumeIcon(this.player.volume);
    }

    updateVolumeIcon(volume) {
        const muteBtn = document.getElementById('muteBtn');
        if (volume === 0) {
            muteBtn.innerHTML = '<i class="bi bi-volume-mute"></i>';
        } else if (volume < 0.5) {
            muteBtn.innerHTML = '<i class="bi bi-volume-down"></i>';
        } else {
            muteBtn.innerHTML = '<i class="bi bi-volume-up"></i>';
        }
    }

    loadChannelInfo() {
        fetch(`/api/channel/detail?channelName=${encodeURIComponent(this.channelName)}`)
            .then(response => {
                if (response.status === 404) {
                    throw new Error('频道不存在');
                }
                return response.json();
            })
            .then(data => {
                if (data.code === 200) {
                    // data.data 直接是数组
                    if (data.data && data.data.length > 0) {
                        // 使用第一个流的信息作为频道信息
                        this.renderChannelInfo(data.data[0]);
                        this.renderStreamList(data.data);
                        // 自动播放第一个流
                        if (data.data[0].StreamUrl && data.data[0].StreamUrl.length > 0) {
                            this.playStream(data.data[0].StreamUrl[0]);
                        }
                    } else {
                        this.showError('暂无可用直播源');
                    }
                } else {
                    this.showError(data.message || '加载失败');
                }
            })
            .catch(error => {
                this.showError(error.message || '加载失败，请稍后重试');
                console.error('Error:', error);
            });
    }

    renderChannelInfo(stream) {
        // 更新频道标题
        document.querySelector('.channel-title').textContent = this.channelName;
    }

    renderStreamList(streams) {
        const streamList = document.getElementById('streamList');
        if (!streams || streams.length === 0) {
            streamList.innerHTML = '<div class="text-center py-4"><p class="text-muted">暂无可用直播源</p></div>';
            return;
        }

        // 展开所有的流URL
        const expandedStreams = streams.flatMap(stream => 
            stream.StreamUrl.map(url => ({
                ...stream,
                singleUrl: url
            }))
        );

        streamList.innerHTML = expandedStreams.map((stream, index) => `
            <div class="list-group-item ${index === 0 ? 'active' : ''}" data-url="${stream.singleUrl}">
                <div class="stream-item">
                    ${stream.StreamLogo ? `
                        <div class="stream-logo">
                            <img src="${stream.StreamLogo}" alt="${stream.StreamName}">
                        </div>
                    ` : ''}
                    <div class="stream-info">
                        <h6 class="stream-name">${stream.StreamName}</h6>
                        <div class="stream-time">
                            更新时间: ${new Date(stream.UpdatedAt * 1000).toLocaleString()}
                        </div>
                    </div>
                </div>
            </div>
        `).join('');

        // 添加事件监听
        streamList.querySelectorAll('.list-group-item').forEach(item => {
            item.addEventListener('click', () => {
                this.playStream(item.dataset.url);
                // 更新选中状态
                streamList.querySelectorAll('.list-group-item').forEach(i => i.classList.remove('active'));
                item.classList.add('active');
            });
        });
    }

    playStream(url) {
        if (this.currentStreamUrl === url) {
            return;
        }
        this.currentStreamUrl = url;

        if (Hls.isSupported()) {
            // 如果存在旧的 HLS 实例，先销毁它
            if (this.hls) {
                this.hls.destroy();
            }
            
            // 创建新的 HLS 实例
            this.hls = new Hls({
                enableWorker: true,
                lowLatencyMode: true
            });

            // 设置事件监听
            this.hls.on(Hls.Events.MANIFEST_PARSED, () => {
                this.player.play().catch(() => {
                    console.log('自动播放被阻止，尝试静音播放');
                    this.player.muted = true;
                    this.player.play();
                });
            });

            // 监听延迟
            this.hls.on(Hls.Events.FRAG_CHANGED, (event, data) => {
                if (data.frag.programDateTime) {
                    const latency = Date.now() - data.frag.programDateTime;
                    document.getElementById('latencyInfo').textContent = `延迟: ${latency}ms`;
                }
            });

            // 加载视频源
            this.hls.loadSource(url);
            this.hls.attachMedia(this.player.media);
        } else if (this.player.media.canPlayType('application/vnd.apple.mpegurl')) {
            // 对于原生支持 HLS 的浏览器（如 Safari）
            this.player.source = {
                type: 'video',
                sources: [{
                    src: url,
                    type: 'application/x-mpegURL'
                }]
            };
            this.player.play().catch(() => {
                console.log('自动播放被阻止，尝试静音播放');
                this.player.muted = true;
                this.player.play();
            });
        }
    }

    copyToClipboard(text) {
        navigator.clipboard.writeText(text).then(() => {
            // 可以添加一个复制成功的提示
            alert('链接已复制到剪贴板');
        }).catch(err => {
            console.error('复制失败:', err);
        });
    }

    showError(message) {
        // 清空频道信息
        document.querySelector('.channel-logo').style.display = 'none';
        document.querySelector('.channel-name').style.display = 'none';
        
        // 显示错误信息
        const channelInfo = document.getElementById('channelInfo');
        channelInfo.innerHTML = `
            <div class="text-center py-5">
                <div class="mb-4">
                    <i class="bi bi-exclamation-circle text-danger" style="font-size: 3rem;"></i>
                </div>
                <h3 class="text-danger mb-3">出错了</h3>
                <p class="text-muted mb-4">${message}</p>
                <a href="${this.channelUrl}" class="btn btn-primary">
                    <i class="bi bi-arrow-left me-2"></i>返回频道列表
                </a>
            </div>
        `;
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    new ChannelDetail();
}); 