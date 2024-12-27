// 存储选中的频道
let selectedChannels = new Set();

// 更新选中计数和按钮状态
function updateSelection() {
    const count = selectedChannels.size;
    document.getElementById('selectedCount').textContent = count;
    document.getElementById('batchActionBtn').disabled = count === 0;
    document.getElementById('verifyBtn').style.display = count > 0 ? 'block' : 'none';
}

// 全选功能
function toggleSelectAll(isChecked) {
    const checkboxes = document.querySelectorAll('.channel-checkbox');
    checkboxes.forEach(checkbox => {
        checkbox.checked = isChecked;
        if (isChecked) {
            selectedChannels.add(checkbox.value);
        } else {
            selectedChannels.delete(checkbox.value);
        }
    });
    updateSelection();
}

// 加载频道列表和记录数
async function loadChannelsWithCount() {
    try {
        // 获取频道列表
        const channelsResponse = await fetch('/api/channels');
        const channelsData = await channelsResponse.json();
        
        if (channelsData.code === 200 && channelsData.data) {
            const channelList = document.getElementById('channelList');
            const channels = channelsData.data;

            // 获取所有频道的记录数
            const recordNumsResponse = await fetch(`/api/channel/get_record_num?${channels.map(ch => `channelName[]=${encodeURIComponent(ch)}`).join('&')}`);
            const recordNumsData = await recordNumsResponse.json();
            const recordNums = recordNumsData.code === 200 ? recordNumsData.data : {};

            channels.forEach(channel => {
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
                    <span class="badge bg-secondary rounded-pill">${recordNums[channel] || 0} 个频道</span>
                `;
                channelList.appendChild(listItem);

                // 添加复选框事件监听
                const checkbox = listItem.querySelector('.channel-checkbox');
                checkbox.addEventListener('change', function() {
                    if (this.checked) {
                        selectedChannels.add(this.value);
                    } else {
                        selectedChannels.delete(this.value);
                    }
                    updateSelection();
                });
            });

            // 添加全选复选框事件监听
            const selectAllCheckbox = document.getElementById('selectAllCheckbox');
            selectAllCheckbox.addEventListener('change', function() {
                toggleSelectAll(this.checked);
            });
        } else {
            console.error('获取频道列表失败:', channelsData.message || '未知错误');
        }
    } catch (error) {
        console.error('请求失败:', error);
    }
}

// 页面加载时执行
loadChannelsWithCount();

// 菜单展开收起功能
const sideMenu = document.querySelector('.side-menu');
const menuToggle = document.querySelector('.menu-toggle');

menuToggle.addEventListener('click', function() {
    sideMenu.classList.toggle('collapsed');
}); 