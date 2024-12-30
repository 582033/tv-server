class ChannelDetailManager {
    constructor() {
        this.channelName = window.location.pathname.split('/').pop();
        this.init();
    }

    init() {
        this.loadStreamData();
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
                            <a href="${url}" target="_blank" class="btn btn-sm btn-outline-success ms-1">
                                <i class="bi bi-play-fill"></i>
                            </a>
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

    copyToClipboard(text) {
        navigator.clipboard.writeText(text)
            .then(() => {
                alert('链接已复制到剪贴板');
            })
            .catch(err => {
                console.error('复制失败:', err);
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