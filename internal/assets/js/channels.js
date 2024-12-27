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

    // 渲染频道列表
    async renderChannelList() {
    try {
        // 获取频道列表
            const channels = await this.fetchChannels();
            if (!channels.length) {
                console.log('没有找到频道');
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
        }
    }

    // 创建频道列表项
    createChannelListItem(channel, recordCount) {
                const listItem = document.createElement('li');
                listItem.className = 'list-group-item d-flex justify-content-between align-items-center';
                listItem.innerHTML = `
                    <div>
                        <input class="form-check-input me-1 channel-checkbox" type="checkbox" 
                               value="${channel}" id="check_${encodeURIComponent(channel)}">
                        <label class="form-check-label channel-name" for="check_${encodeURIComponent(channel)}">
                            ${channel}
                        </label>
                    </div>
            <span class="badge bg-secondary rounded-pill">${recordCount} 个频道</span>
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
    async handleValidate(event) {
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

        // 禁用验证按钮
        verifyBtn.disabled = true;
        verifyBtn.textContent = '验证中...';

        try {
            // 获取延迟设置
            const latency = parseInt(document.getElementById('toleranceSlider').value) || 1000;

            // 先发送验证请求并等待响应
            console.log('开始发送验证请求...');
            const validateResponse = await fetch('/api/channel/validate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify({
                    channelNames: Array.from(this.selectedChannels),
                    timeout: latency
                })
            });

            if (!validateResponse.ok) {
                throw new Error(`验证请求失败: ${validateResponse.status}`);
            }

            console.log('验证请求已发送，开始轮询进度...');

            // 开始轮询进度，直到收到100%或超时
            let startTime = Date.now();
            let timeout = 5 * 60 * 1000; // 5分钟超时
            let lastProgress = 0;

            while (true) {
                const response = await fetch('/api/process');
                const data = await response.json();
                
                if (!data.success) {
                    console.error('进度查询失败:', data);
                } else {
                    const progress = data.process;
                    console.log('当前进度:', progress, '时间:', new Date().toLocaleTimeString());
                    
                    if (progress !== lastProgress) {
                        // 更新进度条
                        this.updateProgressBar(progressBar, progress);
                        lastProgress = progress;
                    }
                    
                    // 检测进度是否到达100%
                    if (progress >= 100) {
                        console.log('进度达到100%，结束轮询');
                        break;
                    }
                }

                // 检查是否超时
                if (Date.now() - startTime > timeout) {
                    console.error('验证超时');
                    throw new Error('验证超时，请稍后重试');
                }

                // 等待1秒后继续轮询
                await new Promise(resolve => setTimeout(resolve, 1000));
            }

            // 获取最终的验证结果
            const validateResult = await validateResponse.json();
            console.log('验证完成，结果:', validateResult);

            // 显示结果
            this.showValidationResult(modalBody, validateResult);
        } catch (error) {
            console.error('验证过程出错:', error);
            this.showValidationResult(modalBody, {
                success: false,
                message: error.message || '验证请求失败，请稍后重试'
            });
        } finally {
            verifyBtn.disabled = false;
            verifyBtn.textContent = '验证';
        }
    }

    // 创建进度条
    createProgressBar(container, verifyBtn) {
        const progressArea = document.createElement('div');
        progressArea.className = 'progress mb-3';
        progressArea.innerHTML = `
            <div class="progress-bar" role="progressbar" style="width: 0%">
                <span class="progress-text">0%</span>
            </div>
        `;
        container.insertBefore(progressArea, verifyBtn);
        return progressArea;
    }

    // 更新进度条
    updateProgressBar(progressBar, progress) {
        progressBar.style.width = `${progress}%`;
        progressBar.querySelector('.progress-text').textContent = `${progress.toFixed(1)}%`;
    }

    // 显示验证结果
    showValidationResult(container, result) {
        const resultDiv = document.createElement('div');
        resultDiv.className = 'alert alert-success mt-3';
        resultDiv.innerHTML = `
            验证完成<br>
            验证频道：${result.stats.total} 个<br>
            验证通过：${result.stats.valid} 个
        `;
        container.appendChild(resultDiv);
    }
}

// 页面加载完成后初始化
window.addEventListener('load', () => {
    const channelManager = new ChannelManager();
    channelManager.renderChannelList();
}); 
