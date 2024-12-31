class ChannelDetailManager {
    constructor() {
        this.channelName = window.location.pathname.split('/').pop();
        this.player = null;
        this.hls = null;
        this.playerModal = null;
        this.init();
    }

    init() {
        this.initPlayer();
        this.loadStreamData();
    }

    initPlayer() {
        // 初始化播放器
        const video = document.getElementById('player');
        this.player = new Plyr(video, {
            controls: [], // 禁用 Plyr 默认控件
            clickToPlay: false, // 禁用点击播放
            keyboard: { focused: true, global: true },
            autoplay: true,
            muted: true,  // 初始静音以确保自动播放
            hideControls: true, // 始终隐藏默认控件
            disableContextMenu: true // 禁用右键菜单
        });

        // 初始化模态框
        this.playerModal = new bootstrap.Modal(document.getElementById('playerModal'));
        
        // 初始化自定义控件
        this.initCustomControls();

        // 监听模态框关闭事件
        document.getElementById('playerModal').addEventListener('hidden.bs.modal', () => {
            if (this.hls) {
                this.hls.destroy();
                this.hls = null;
            }
            if (this.latencyInterval) {
                clearInterval(this.latencyInterval);
                this.latencyInterval = null;
            }
            this.player.stop();
        });

        // 监听播放器就绪事件
        this.player.on('ready', () => {
            this.player.volume = 0.5;
            document.getElementById('volumeSlider').value = this.player.volume * 100;
        });
    }

    initCustomControls() {
        // 播放/暂停按钮
        const playPauseBtn = document.getElementById('playPauseBtn');
        playPauseBtn.addEventListener('click', () => {
            if (this.player.playing) {
                this.player.pause();
                playPauseBtn.innerHTML = '<i class="bi bi-play-fill"></i>';
            } else {
                this.player.play();
                playPauseBtn.innerHTML = '<i class="bi bi-pause-fill"></i>';
            }
        });

        // 进度条
        const progressBar = document.getElementById('progressBar');
        const progress = progressBar.parentElement;
        
        progress.addEventListener('click', (e) => {
            const rect = progress.getBoundingClientRect();
            const pos = (e.clientX - rect.left) / rect.width;
            this.player.currentTime = pos * this.player.duration;
        });

        // 时间更新
        this.player.on('timeupdate', () => {
            const currentTime = document.getElementById('currentTime');
            const duration = document.getElementById('duration');
            
            currentTime.textContent = this.formatPlayerTime(this.player.currentTime);
            duration.textContent = this.formatPlayerTime(this.player.duration);
            
            const progress = (this.player.currentTime / this.player.duration) * 100;
            progressBar.style.width = `${progress}%`;
        });

        // 音量控制
        const volumeSlider = document.getElementById('volumeSlider');
        const muteBtn = document.getElementById('muteBtn');

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
        const fullscreenBtn = document.getElementById('fullscreenBtn');
        fullscreenBtn.addEventListener('click', () => {
            this.player.fullscreen.toggle();
        });

        // 播放状态监听
        this.player.on('play', () => {
            playPauseBtn.innerHTML = '<i class="bi bi-pause-fill"></i>';
        });

        this.player.on('pause', () => {
            playPauseBtn.innerHTML = '<i class="bi bi-play-fill"></i>';
        });

        // 全屏状态监听
        this.player.on('enterfullscreen', () => {
            fullscreenBtn.innerHTML = '<i class="bi bi-fullscreen-exit"></i>';
        });

        this.player.on('exitfullscreen', () => {
            fullscreenBtn.innerHTML = '<i class="bi bi-fullscreen"></i>';
        });
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

    formatPlayerTime(seconds) {
        if (isNaN(seconds)) return '00:00';
        const mins = Math.floor(seconds / 60);
        const secs = Math.floor(seconds % 60);
        return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    }

    loadStreamData() {
        // 显示加载动画
        document.getElementById('loading').classList.remove('d-none');
        document.getElementById('streamTable').classList.add('d-none');
        document.getElementById('errorMessage').classList.add('d-none');
        document.getElementById('emptyMessage').classList.add('d-none');

        // 获取频道详情数据
        fetch(`/api/channel/detail?channelName=${encodeURIComponent(this.channelName)}`)
            .then(response => response.json())
            .then(data => {
                if (data.code === 200 && data.data) {
                    this.renderStreamData(data.data);
                } else {
                    throw new Error(data.message || '加载失败');
                }
            })
            .catch(error => {
                console.error('加载频道详情失败:', error);
                document.getElementById('errorMessage').classList.remove('d-none');
            })
            .finally(() => {
                document.getElementById('loading').classList.add('d-none');
            });
    }

    renderStreamData(streams) {
        if (!streams || streams.length === 0) {
            document.getElementById('emptyMessage').classList.remove('d-none');
            return;
        }

        const tbody = document.getElementById('streamTableBody');
        tbody.innerHTML = '';

        streams.forEach(stream => {
            const tr = document.createElement('tr');
            tr.innerHTML = this.createStreamRow(stream);
            tbody.appendChild(tr);
        });

        document.getElementById('streamTable').classList.remove('d-none');
    }

    createStreamRow(stream) {
        return `
            <td>${stream.StreamName}</td>
            <td>
                ${stream.StreamLogo ? 
                    `<img src="${stream.StreamLogo}" alt="Logo" style="height: 30px; width: auto;">` : 
                    '<span class="text-muted">-</span>'}
            </td>
            <td>
                <div class="stream-urls">
                    ${stream.StreamUrl.map(url => `
                        <div class="d-flex align-items-center mb-1">
                            <span class="text-truncate" style="max-width: 300px;">${url}</span>
                            <button class="btn btn-sm btn-outline-primary ms-2" onclick="channelDetail.copyToClipboard('${url}')">
                                <i class="bi bi-clipboard"></i>
                            </button>
                            <button class="btn btn-sm btn-outline-success ms-1" onclick="channelDetail.playStream('${url}', '${stream.StreamName}')">
                                <i class="bi bi-play-fill"></i>
                            </button>
                        </div>
                    `).join('')}
                </div>
            </td>
            <td>${this.formatTime(stream.CreatedAt)}</td>
            <td>${this.formatTime(stream.UpdatedAt)}</td>
            <td>
                <div class="btn-group">
                    <button class="btn btn-sm btn-outline-primary" onclick="channelDetail.validateStream('${stream.ID}')">
                        <i class="bi bi-check2-circle"></i> 验证
                    </button>
                    <button class="btn btn-sm btn-outline-danger" onclick="channelDetail.deleteStream('${stream.ID}')">
                        <i class="bi bi-trash"></i> 删除
                    </button>
                </div>
            </td>
        `;
    }

    playStream(url, title) {
        // 设置标题
        document.getElementById('playerTitle').textContent = title;

        // 如果存在旧的HLS实例，销毁它
        if (this.hls) {
            this.hls.destroy();
            this.hls = null;
        }

        // 显示模态框
        this.playerModal.show();

        // 创建新的HLS实例
        if (Hls.isSupported()) {
            this.hls = new Hls({
                enableWorker: true,
                lowLatencyMode: true,
                debug: false,
                // 添加延迟优化配置
                liveSyncDurationCount: 3,    // 直播同步时间片数量
                liveMaxLatencyDurationCount: 5, // 最大延迟时间片数量
                liveDurationInfinity: true,   // 无限时长模式
                highBufferWatchdogPeriod: 1,  // 缓冲区监控周期
                autoStartLoad: true,          // 自动开始加载
                startLevel: -1,               // 自动选择初始质量
                defaultAudioCodec: undefined  // 自动选择音频编码
            });

            this.hls.loadSource(url);
            this.hls.attachMedia(this.player.media);

            // 监听事件以更新延迟信息
            this.hls.on(Hls.Events.MANIFEST_PARSED, () => {
                this.player.play().catch(() => {
                    console.log('自动播放被阻止，尝试静音播放');
                    this.player.muted = true;
                    this.player.play();
                });
            });

            // 监听延迟更新
            this.hls.on(Hls.Events.FRAG_CHANGED, (event, data) => {
                this.updateLatency();
            });

            // 定期更新延迟信息
            this.latencyInterval = setInterval(() => this.updateLatency(), 1000);
        } else if (this.player.media.canPlayType('application/vnd.apple.mpegurl')) {
            // 对于Safari等原生支持HLS的浏览器
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

    updateLatency() {
        if (!this.hls || !this.hls.media) return;

        const latency = this.calculateLatency();
        const latencyElement = document.getElementById('latencyValue');
        
        if (latencyElement && !isNaN(latency)) {
            // 根据延迟值设置不同的颜色
            const badge = document.getElementById('streamStats');
            if (badge) {
                if (latency < 5) {
                    badge.className = 'badge bg-success';
                } else if (latency < 10) {
                    badge.className = 'badge bg-warning';
                } else {
                    badge.className = 'badge bg-danger';
                }
            }
            
            latencyElement.textContent = `${latency.toFixed(2)}秒`;
        }
    }

    calculateLatency() {
        if (!this.hls || !this.hls.media) return 0;

        const liveEdge = this.hls.liveSyncPosition;
        const currentTime = this.hls.media.currentTime;
        
        if (liveEdge === null || currentTime === 0) return 0;
        
        return liveEdge - currentTime;
    }

    copyToClipboard(text) {
        navigator.clipboard.writeText(text)
            .then(() => {
                // 使用 Bootstrap 的 Toast 提示
                const toast = document.createElement('div');
                toast.className = 'position-fixed bottom-0 end-0 p-3';
                toast.style.zIndex = '5000';
                toast.innerHTML = `
                    <div class="toast align-items-center text-white bg-success border-0" role="alert">
                        <div class="d-flex">
                            <div class="toast-body">
                                <i class="bi bi-check-circle me-2"></i>链接已复制到剪贴板
                            </div>
                            <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
                        </div>
                    </div>
                `;
                document.body.appendChild(toast);
                const bsToast = new bootstrap.Toast(toast.querySelector('.toast'));
                bsToast.show();
                toast.addEventListener('hidden.bs.toast', () => {
                    document.body.removeChild(toast);
                });
            })
            .catch(err => {
                console.error('复制失败:', err);
                alert('复制失败: ' + err.message);
            });
    }

    formatTime(timestamp) {
        const date = new Date(timestamp * 1000);
        return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    }

    validateStream(id) {
        if (!confirm('确定要验证这个媒体流吗？')) {
            return;
        }

        fetch('/api/channel/validate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                streamId: id
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.code === 200) {
                alert('验证成功');
                this.loadStreamData(); // 重新加载数据
            } else {
                throw new Error(data.message || '验证失败');
            }
        })
        .catch(error => {
            console.error('验证失败:', error);
            alert('验证失败: ' + error.message);
        });
    }

    deleteStream(id) {
        if (!confirm('确定要删除这个媒体流吗？')) {
            return;
        }

        fetch('/api/channel/delete', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                streamId: id
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.code === 200) {
                alert('删除成功');
                this.loadStreamData(); // 重新加载数据
            } else {
                throw new Error(data.message || '删除失败');
            }
        })
        .catch(error => {
            console.error('删除失败:', error);
            alert('删除失败: ' + error.message);
        });
    }
}

// 初始化
const channelDetail = new ChannelDetailManager(); 