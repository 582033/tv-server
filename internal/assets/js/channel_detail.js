class ChannelDetail {
    constructor() {
        this.player = null;
        this.channelName = window.CHANNEL_NAME;
        this.channelUrl = window.CHANNEL_URL;
        this.latencyUpdateInterval = null;
        
        // 延迟计算相关
        this.latencySamples = [];
        this.MAX_LATENCY_SAMPLES = 5;
        
        this.initPlayer();
        this.loadChannelInfo();

        // 在页面卸载时清理定时器和播放器
        window.addEventListener('beforeunload', () => {
            if (this.latencyUpdateInterval) {
                clearInterval(this.latencyUpdateInterval);
            }
            if (this.player) {
                this.player.dispose();
            }
        });
    }

    initPlayer() {
        // 初始化 video.js 播放器
        this.player = videojs('player', {
            fluid: true,
            aspectRatio: '16:9',
            autoplay: true,      // 改为直接启用自动播放
            muted: true,         // 初始静音以确保自动播放
            controls: true,
            preload: 'auto',
            playbackRates: [0.5, 1, 1.5, 2],
            playsinline: true,   // 禁用自动全屏播放
            webkitPlaysinline: true,
            controlBar: {
                children: [
                    'playToggle',
                    'volumePanel',
                    'currentTimeDisplay',
                    'timeDivider',
                    'durationDisplay',
                    'progressControl',
                    'liveDisplay',
                    'customControlSpacer',
                    'playbackRateMenuButton',
                    'fullscreenToggle'
                ]
            },
            html5: {
                vhs: {
                    overrideNative: true,
                    enableLowInitialPlaylist: true,
                    smoothQualityChange: true,
                    fastQualityChange: true
                },
                nativeAudioTracks: false,
                nativeVideoTracks: false
            }
        });

        // 添加内联播放属性到视频元素
        const videoElement = document.querySelector('#player_html5_api');
        if (videoElement) {
            videoElement.setAttribute('playsinline', 'true');
            videoElement.setAttribute('webkit-playsinline', 'true');
            videoElement.setAttribute('x5-playsinline', 'true');
            videoElement.setAttribute('x5-video-player-type', 'h5');
            videoElement.setAttribute('x5-video-player-fullscreen', 'false');
        }

        // 监听播放器事件
        this.player.on('loadedmetadata', () => {
            // 视频元数据加载完成后强制播放
            const playAttempt = this.player.play();
            if (playAttempt !== undefined) {
                playAttempt.catch(error => {
                    console.log('自动播放失败，尝试静音播放:', error);
                    this.player.muted(true);
                    this.player.play();
                });
            }
        });

        // 监听暂停事件，防止自动暂停
        this.player.on('pause', () => {
            // 如果是第一次加载的视频，则立即恢复播放
            if (this.isFirstPlay) {
                this.player.play();
            }
        });

        this.player.on('playing', () => {
            // 如果是静音状态，等待一段时间后尝试取消静音
            if (this.player.muted()) {
                setTimeout(() => {
                    this.player.muted(false);
                    this.player.volume(0.6);
                }, 1000);
            }
        });

        // 监听全屏变化事件
        this.player.on('fullscreenchange', () => {
            const isFullscreen = this.player.isFullscreen();
            const playerContainer = document.getElementById('playerContainer');
            
            if (isFullscreen) {
                playerContainer.classList.add('fullscreen');
            } else {
                playerContainer.classList.remove('fullscreen');
                // 强制重新计算布局
                playerContainer.style.height = window.innerWidth <= 768 ? '40vh' : 'calc(100vh - 140px)';
            }
        });

        // 监听错误事件
        this.player.on('error', (error) => {
            console.error('播放器错误:', error);
        });

        // 设置首次播放标志
        this.isFirstPlay = true;
    }

    playStream(url) {
        if (!this.player || !url) return;

        const latencyInfo = document.getElementById('latencyInfo');
        if (latencyInfo) {
            latencyInfo.textContent = '计算中';
        }

        // 清理旧的定时器
        if (this.latencyUpdateInterval) {
            clearInterval(this.latencyUpdateInterval);
        }

        // 立即更新一次延迟
        this.updatePlayerLatency(url);
        
        // 设置定期更新
        this.latencyUpdateInterval = setInterval(() => {
            this.updatePlayerLatency(url);
        }, 5000);

        // 确保静音状态以支持自动播放
        this.player.muted(true);

        // 设置新的播放源并立即播放
        this.player.src({
            src: url,
            type: 'application/x-mpegURL'
        });

        // 强制开始播放
        const playPromise = this.player.play();
        
        if (playPromise !== undefined) {
            playPromise.then(() => {
                // 播放成功后，等待一段时间再尝试取消静音
                setTimeout(() => {
                    this.player.muted(false);
                    this.player.volume(0.6);
                }, 1000);
            }).catch(error => {
                console.error('播放失败:', error);
                // 如果普通播放失败，保持静音状态继续播放
                this.player.muted(true);
                this.player.play().catch(e => {
                    console.error('静音播放也失败:', e);
                });
            });
        }

        // 如果不是第一个视频，则取消首次播放标志
        if (!this.isFirstPlay) {
            this.isFirstPlay = false;
        }
    }

    // 测试单个流的延迟
    testStreamLatency(url) {
        return new Promise((resolve) => {
            const startTime = performance.now();
            const xhr = new XMLHttpRequest();
            
            xhr.open('GET', url, true);
            xhr.timeout = 5000;  // 5秒超时
            
            xhr.onload = function() {
                if (xhr.status >= 200 && xhr.status < 300) {
                    // 获取请求的性能数据
                    const entries = performance.getEntriesByType('resource')
                        .filter(entry => entry.name === url);
                    
                    if (entries.length > 0) {
                        // 使用最新的请求记录
                        const entry = entries[entries.length - 1];
                        // 计算总耗时（包括 DNS、TCP、请求、响应等全部时间）
                        const latency = Math.round(entry.duration);
                        resolve({ url, latency: latency, error: null });
                    } else {
                        // 如果找不到性能数据，使用简单的时间差
                        const latency = Math.round(performance.now() - startTime);
                        resolve({ url, latency: latency, error: null });
                    }
                } else {
                    resolve({ url, latency: null, error: '加载失败' });
                }
            };
            
            xhr.onerror = function() {
                resolve({ url, latency: null, error: '加载失败' });
            };
            
            xhr.ontimeout = function() {
                resolve({ url, latency: null, error: '超时' });
            };
            
            xhr.send();
            performance.clearResourceTimings();
        });
    }

    // 更新播放器延迟显示
    updatePlayerLatency(url) {
        const latencyInfo = document.getElementById('latencyInfo');
        if (!latencyInfo) return;

        this.testStreamLatency(url).then(result => {
            if (result.error) {
                latencyInfo.textContent = result.error;
            } else if (result.latency !== null) {
                latencyInfo.textContent = `${result.latency}ms`;
            } else {
                latencyInfo.textContent = '计算中';
            }
        }).catch(() => {
            latencyInfo.textContent = '计算失败';
        });
    }

    // 测试所有流的延迟
    testStreamLatencies(streams) {
        const batchSize = 5; // 同时测试的流数量
        const updateLatencyDisplay = (url, result) => {
            const latencyElement = document.querySelector(`.stream-latency[data-url="${url}"]`);
            if (latencyElement) {
                if (result.error) {
                    latencyElement.innerHTML = `<span class="badge bg-danger">${result.error}</span>`;
                } else if (result.latency !== null) {
                    const latencyMs = Math.round(result.latency);
                    let badgeClass = 'bg-success';
                    if (latencyMs > 1000) {
                        badgeClass = 'bg-danger';
                    } else if (latencyMs > 500) {
                        badgeClass = 'bg-warning';
                    }
                    latencyElement.innerHTML = `<span class="badge ${badgeClass}">${latencyMs}ms</span>`;
                } else {
                    latencyElement.innerHTML = `<span class="badge bg-secondary">未知</span>`;
                }
            }
        };

        // 递归处理每个批次
        const processBatch = (startIndex) => {
            if (startIndex >= streams.length) return;

            const batch = streams.slice(startIndex, startIndex + batchSize);
            const promises = batch.map(stream => 
                new Promise(resolve => {
                    setTimeout(() => {
                        this.testStreamLatency(stream.singleUrl)
                            .then(resolve);
                    }, Math.random() * 500);
                })
            );

            Promise.all(promises)
                .then(results => {
                    results.forEach(result => {
                        updateLatencyDisplay(result.url, result);
                    });
                    setTimeout(() => {
                        processBatch(startIndex + batchSize);
                    }, 200);
                })
                .catch(error => {
                    console.error('测试延迟时出错:', error);
                    setTimeout(() => {
                        processBatch(startIndex + batchSize);
                    }, 200);
                });
        };

        processBatch(0);
    }

    // 显示错误信息
    showError(message) {
        const channelInfo = document.getElementById('channelInfo');
        if (channelInfo) {
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

    // 加载频道信息
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
                    if (data.data && data.data.length > 0) {
                        this.renderChannelInfo(data.data[0]);
                        this.renderStreamList(data.data);
                        
                        // 确保播放器已初始化
                        if (this.player && data.data[0].streamUrl && data.data[0].streamUrl.length > 0) {
                            // 设置一个短暂延迟确保DOM完全加载
                            setTimeout(() => {
                                this.playStream(data.data[0].streamUrl[0]);
                            }, 100);
                        }
                    } else {
                        this.showError('暂无可用直播源');
                    }
                } else {
                    this.showError(data.message || '加载失败');
                }
            })
            .catch(error => {
                console.error('加载出错:', error);
                this.showError(error.message || '加载失败，请稍后重试');
            });
    }

    // 渲染频道信息
    renderChannelInfo(stream) {
        const titleElement = document.querySelector('.channel-title');
        if (titleElement) {
            titleElement.textContent = this.channelName;
        }
    }

    // 渲染流列表
    renderStreamList(streams) {
        const streamList = document.getElementById('streamList');
        if (!streamList) return;

        if (!streams || streams.length === 0) {
            streamList.innerHTML = '<div class="text-center py-4"><p class="text-muted">暂无可用直播源</p></div>';
            return;
        }

        // 展开所有的流URL
        const expandedStreams = streams.flatMap(stream => {
            const urls = Array.isArray(stream.streamUrl) ? stream.streamUrl : [];
            return urls.map(url => ({
                ...stream,
                singleUrl: url
            }));
        });

        if (expandedStreams.length === 0) {
            streamList.innerHTML = '<div class="text-center py-4"><p class="text-muted">暂无可用直播源</p></div>';
            return;
        }

        streamList.innerHTML = expandedStreams.map((stream, index) => `
            <div class="list-group-item ${index === 0 ? 'active' : ''}" data-url="${stream.singleUrl}">
                <div class="stream-item">
                    ${stream.streamLogo ? `
                        <div class="stream-logo">
                            <img src="${stream.streamLogo}">
                        </div>
                    ` : ''}
                    <div class="stream-info">
                        <div class="stream-header">
                            <h6 class="stream-name mb-0">${stream.streamName || '未知频道'}</h6>
                            <button class="favorite-btn" title="收藏频道">
                                <i class="bi bi-heart"></i>
                            </button>
                        </div>
                        <div class="stream-footer">
                            <div class="stream-time">
                                更新时间: ${stream.updatedAt ? new Date(stream.updatedAt * 1000).toLocaleString() : '未知'}
                            </div>
                            <div class="stream-latency" data-url="${stream.singleUrl}">
                                <span class="badge bg-secondary">延迟测试中...</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `).join('');

        // 添加点击事件
        streamList.querySelectorAll('.list-group-item').forEach(item => {
            item.addEventListener('click', (e) => {
                if (e.target.closest('.favorite-btn')) {
                    e.preventDefault();
                    e.stopPropagation();
                    const btn = e.target.closest('.favorite-btn');
                    btn.classList.toggle('active');
                    btn.querySelector('i').classList.toggle('bi-heart');
                    btn.querySelector('i').classList.toggle('bi-heart-fill');
                    return;
                }

                const url = item.dataset.url;
                if (url) {
                    this.playStream(url);
                    streamList.querySelectorAll('.list-group-item').forEach(i => i.classList.remove('active'));
                    item.classList.add('active');
                }
            });
        });

        // 测试延迟
        if (expandedStreams.length > 0) {
            this.testStreamLatencies(expandedStreams);
        }
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    new ChannelDetail();
}); 