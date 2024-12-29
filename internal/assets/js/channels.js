// 频道管理类
class ChannelManager {
    constructor() {
        this.selectedChannels = new Set();
        this.initializeEventListeners();
    }

    // 初始化所有事件监听器
    initializeEventListeners() {
        // 全选复选框
        const selectAllCheckbox = document.getElementById('selectAllCheckbox');
        if (selectAllCheckbox) {
            selectAllCheckbox.addEventListener('change', () => this.handleSelectAll(selectAllCheckbox));
        }

        // 模态框中的验证按钮
        const verifyBtn = document.querySelector('.modal-body .verify-btn');
        if (verifyBtn) {
            verifyBtn.addEventListener('click', (e) => this.handleValidate(e));
        }
    }

    // 更新选中状态和UI
    updateSelectionStatus() {
        const count = this.selectedChannels.size;
        
        // 更新计数
        const countElement = document.getElementById('selectedCount');
        if (countElement) {
            countElement.textContent = count;
        }

        // 更新批量操作按钮状态
        const batchBtn = document.getElementById('batchActionBtn');
        if (batchBtn) {
            batchBtn.disabled = count === 0;
        }

        // 更新验证按钮显示状态
        const verifyBtn = document.getElementById('verifyBtn');
        if (verifyBtn) {
            verifyBtn.style.display = count > 0 ? 'block' : 'none';
        }
    }

    // 处理全选/取消全选
    handleSelectAll(checkbox) {
        const channelCheckboxes = document.querySelectorAll('.channel-checkbox');
        channelCheckboxes.forEach(item => {
            item.checked = checkbox.checked;
            if (checkbox.checked) {
                this.selectedChannels.add(item.value);
            } else {
                this.selectedChannels.delete(item.value);
            }
        });
        this.updateSelectionStatus();
    }

    // 处理单个频道选择
    handleChannelSelect(checkbox) {
        if (checkbox.checked) {
            this.selectedChannels.add(checkbox.value);
        } else {
            this.selectedChannels.delete(checkbox.value);
            // 取消全选复选框
            const selectAllCheckbox = document.getElementById('selectAllCheckbox');
            if (selectAllCheckbox) {
                selectAllCheckbox.checked = false;
            }
        }
        this.updateSelectionStatus();
    }

    // 显示loading动画
    showLoading() {
        const channelList = document.getElementById('channelList');
        if (!channelList) return;

        channelList.innerHTML = `
            <div class="text-center p-5">
                <div class="spinner-border text-primary" role="status">
                    <span class="visually-hidden">加载中...</span>
                </div>
                <div class="mt-2">加载中...</div>
            </div>
        `;
    }

    // 渲染频道列表
    async renderChannelList() {
        try {
            // 显示loading动画
            this.showLoading();
            
        // 获取频道列表
            const channels = await this.fetchChannels();
            if (!channels.length) {
                // 隐藏loading动画
                this.hideLoading();
                // 显示错误信息
                const channelList = document.getElementById('channelList');
                if (channelList) {
                    channelList.innerHTML = `
                        <div class="alert alert-danger m-3" role="alert">
                            没有找到频道
                        </div>
                    `;
                }
                return;
            }

            // 获取频道记录数
            const recordNums = await this.fetchChannelRecords(channels);

            // 渲染列表
            const channelList = document.getElementById('channelList');
            if (!channelList) return;

            // 清空现有内容
            channelList.innerHTML = '';
            this.selectedChannels.clear();

            // 创建并添加列表项
            channels.forEach(channel => {
                const listItem = this.createChannelListItem(channel, recordNums[channel] || 0);
                channelList.appendChild(listItem);
            });

            // 重置全选复选框
            const selectAllCheckbox = document.getElementById('selectAllCheckbox');
            if (selectAllCheckbox) {
                selectAllCheckbox.checked = false;
            }

            this.updateSelectionStatus();
        } catch (error) {
            console.error('渲染频道列表失败:', error);
            // 显示错误信息
            const channelList = document.getElementById('channelList');
            if (channelList) {
                channelList.innerHTML = `
                    <div class="alert alert-danger m-3" role="alert">
                        加载频道列表失败，请稍后重试
                    </div>
                `;
            }
        }
    }

    // 创建频道列表项
    createChannelListItem(channel, recordCount) {
                const listItem = document.createElement('li');
        listItem.className = 'list-group-item';
                listItem.innerHTML = `
            <div class="row align-items-center">
                <div class="col-auto">
                    <input class="form-check-input channel-checkbox" type="checkbox" 
                               value="${channel}" id="check_${encodeURIComponent(channel)}">
                </div>
                <div class="col">
                    <div class="d-flex justify-content-between align-items-center">
                        <label class="form-check-label channel-name mb-0" for="check_${encodeURIComponent(channel)}">
                            ${channel}
                        </label>
                        <div class="d-flex align-items-center">
                            <span class="badge bg-secondary rounded-pill">${recordCount} 个频道</span>
                            <a href="/channel/${encodeURIComponent(channel)}" class="btn btn-link text-decoration-none p-0 ms-3">
                                详情 <i class="bi bi-chevron-right"></i>
                            </a>
                        </div>
                    </div>
                </div>
            </div>
                `;

                // 添加复选框事件监听
                const checkbox = listItem.querySelector('.channel-checkbox');
        if (checkbox) {
            checkbox.addEventListener('change', () => this.handleChannelSelect(checkbox));
        }

        return listItem;
    }

    // 获取频道列表
    async fetchChannels() {
        try {
            const response = await fetch('/api/channels');
            const data = await response.json();
            if (data.code === 200 && data.data) {
                return data.data;
            }
            throw new Error(data.message || '获取频道列表失败');
        } catch (error) {
            console.error('获取频道列表失败:', error);
            return [];
        }
    }

    // 获取频道记录数
    async fetchChannelRecords(channels) {
        try {
            const queryString = channels.map(ch => `channelName[]=${encodeURIComponent(ch)}`).join('&');
            const response = await fetch(`/api/channel/get_record_num?${queryString}`);
            const data = await response.json();
            return data.code === 200 ? data.data : {};
        } catch (error) {
            console.error('获取频道记录数失败:', error);
            return {};
        }
    }

    // 处理验证请求
    handleValidate(event) {
        event.preventDefault();

        if (this.selectedChannels.size === 0) {
            alert('请先选择要验证的频道');
            return;
        }

        const modalBody = document.querySelector('.modal-body');
        const verifyBtn = event.target;
        if (!modalBody || !verifyBtn) return;

        // 清除旧的进度条和结果
        modalBody.querySelectorAll('.progress, .alert').forEach(el => el.remove());

        // 创建进度条并插入到验证按钮之前
        const progressArea = this.createProgressBar(modalBody, verifyBtn);
        const progressBar = progressArea.querySelector('.progress-bar');
        const progressText = progressArea.querySelector('.progress-text');

        // 禁用验证按钮
        verifyBtn.disabled = true;
        verifyBtn.textContent = '验证中...';

        // 开始轮询进度
        let progressInterval = setInterval(() => {
            fetch('/api/process')
                .then(response => response.json())
                .then(result => {
                    if (result.code === 200 && result.data) {
                        const progress = result.data.process;
                        progressBar.style.width = `${progress}%`;
                        progressText.textContent = `${progress.toFixed(1)}%`;
                        
                        // 如果进度达到100%，停止轮询
                        if (progress >= 100) {
                            clearInterval(progressInterval);
                        }
                    }
                })
                .catch(error => {
                    console.error('获取进度失败:', error);
                });
        }, 1000); // 每秒更新一次进度

        // 发送验证请求
        fetch('/api/channel/validate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                channelNames: Array.from(this.selectedChannels),
                timeout: parseInt(document.getElementById('toleranceSlider').value) || 1000
            })
        })
        .then(response => response.json())
        .then(result => {
            // 创建结果显示区域
            const resultDiv = document.createElement('div');
            resultDiv.className = `alert ${result.success ? 'alert-success' : 'alert-danger'} mt-3`;
            
            if (result.success) {
                resultDiv.innerHTML = `
                    原始链接：${result.stats.total} 个<br>
                    验证通过：${result.stats.unique} 个<br>
                    有效链接：${result.stats.valid} 个<br>`;

                if (result.stats.valid > 0) {
                    resultDiv.innerHTML += `
                        您可以通过以下地址访问合并后的 M3U 文件：
                        <a href="${result.m3uLink}" target="_blank">${result.m3uLink}</a>
                    `;
                }
            } else {
                resultDiv.textContent = result.message || '验证失败';
            }
            
            modalBody.appendChild(resultDiv);
        })
        .catch(error => {
            console.error('验证过程出错:', error);
            const errorDiv = document.createElement('div');
            errorDiv.className = 'alert alert-danger mt-3';
            errorDiv.textContent = error.message || '验证请求失败，请稍后重试';
            modalBody.appendChild(errorDiv);
        })
        .finally(() => {
            verifyBtn.disabled = false;
            verifyBtn.textContent = '验证';
            clearInterval(progressInterval);
            progressArea.remove();
        });
    }

    // 创建进度条
    createProgressBar(container, verifyBtn) {
        const progressArea = document.createElement('div');
        progressArea.className = 'progress mb-3';
        progressArea.style.height = '20px';
        progressArea.style.position = 'relative';
        
        const progressBar = document.createElement('div');
        progressBar.className = 'progress-bar progress-bar-striped progress-bar-animated bg-primary';
        progressBar.setAttribute('role', 'progressbar');
        progressBar.style.width = '0%';
        
        const progressText = document.createElement('span');
        progressText.className = 'progress-text';
        progressText.style.position = 'absolute';
        progressText.style.left = '50%';
        progressText.style.top = '50%';
        progressText.style.transform = 'translate(-50%, -50%)';
        progressText.style.color = '#fff';
        progressText.style.zIndex = '1';
        progressText.textContent = '0%';
        
        progressArea.appendChild(progressBar);
        progressArea.appendChild(progressText);
        container.insertBefore(progressArea, verifyBtn);
        
        return progressArea;
    }

    // 更新进度条
    updateProgressBar(progressArea, progress) {
        console.log('开始更新进度条，当前进度值:', progress);
        
        if (!progressArea) {
            console.error('进度条容器不存在');
            return;
        }

        // 获取进度条和文本元素
        const progressBar = progressArea.querySelector('.progress-bar');
        const progressText = progressArea.querySelector('.progress-text');

        if (!progressBar || !progressText) {
            console.error('找不到进度条或文本元素:', {
                progressBar: !!progressBar,
                progressText: !!progressText
            });
            return;
        }

        // 确保进度是数字
        const progressValue = parseFloat(progress);
        if (isNaN(progressValue)) {
            console.error('进度值无效:', progress);
            return;
        }

        console.log('更新前状态:', {
            width: progressBar.style.width,
            text: progressText.textContent
        });

        // 更新进度条宽度
        const widthValue = progressValue + '%';
        progressBar.style.width = widthValue;
        
        // 更新文本
        const textValue = progressValue.toFixed(1) + '%';
        progressText.textContent = textValue;

        console.log('更新后状态:', {
            width: progressBar.style.width,
            text: progressText.textContent
        });
    }

    // 显示验证结果
    showValidationResult(container, result) {
        const resultDiv = document.createElement('div');
        resultDiv.className = `alert ${result.success ? 'alert-success' : 'alert-danger'} mt-3`;
        
        if (result.success && result.stats) {
            resultDiv.innerHTML = `
                验证完成<br>
                验证频道：${result.stats.total || 0} 个<br>
                验证通过：${result.stats.valid || 0} 个
            `;
        } else {
            resultDiv.textContent = result.message || '验证失败';
        }
        
        container.appendChild(resultDiv);
    }
}

// 页面加载完成后初始化
window.addEventListener('load', () => {
    const channelManager = new ChannelManager();
    channelManager.renderChannelList();
}); 
