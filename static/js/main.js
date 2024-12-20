// 显示结果提示
function showResult(type, message) {
    const result = document.getElementById('result');
    result.innerHTML = `
        <div class="alert alert-${type === 'success' ? 'success' : 'danger'}" role="alert">
            ${message}
        </div>
    `;
}

document.addEventListener('DOMContentLoaded', function() {
    const urlInputs = document.getElementById('urlInputs');
    const addUrlBtn = document.getElementById('addUrlBtn');
    const validateBtn = document.getElementById('validateBtn');
    const template = document.getElementById('urlInputTemplate');

    // 添加链接输入框
    function addUrlInput(value = '') {
        const newInput = template.content.cloneNode(true);
        urlInputs.appendChild(newInput);
        
        const input = newInput.querySelector('input');
        const removeBtn = newInput.querySelector('.btn-remove');

        input.addEventListener('input', validateInput);
        removeBtn.addEventListener('click', () => removeUrlInput(newInput));

        if (value) {
            input.value = value;
        }

        validateAllInputs();
        return newInput;
    }

    // 删除链接输入框
    function removeUrlInput(inputGroup) {
        if (urlInputs.children.length > 1) {
            inputGroup.classList.add('removing');
            setTimeout(() => {
                inputGroup.remove();
                validateAllInputs();
            }, 300);
        }
    }

    // 验证单个输入框
    function validateInput(event) {
        const input = event.target;
        const inputGroup = input.closest('.url-input-group');
        const url = input.value.trim();

        try {
            const urlObj = new URL(url);
            const isValid = (urlObj.protocol === 'http:' || urlObj.protocol === 'https:') 
                && (url.toLowerCase().endsWith('.m3u') || url.toLowerCase().endsWith('.m3u8'));
            
            inputGroup.classList.toggle('is-invalid', !isValid && url !== '');
            validateBtn.disabled = !isValid;
            return isValid;
        } catch {
            inputGroup.classList.toggle('is-invalid', url !== '');
            validateBtn.disabled = true;
            return false;
        }
    }

    // 验证所有输入框
    function validateAllInputs() {
        const inputs = urlInputs.querySelectorAll('input');
        const validUrls = Array.from(inputs)
            .map(input => input.value.trim())
            .filter(url => {
                try {
                    const urlObj = new URL(url);
                    return (urlObj.protocol === 'http:' || urlObj.protocol === 'https:') 
                        && (url.toLowerCase().endsWith('.m3u') || url.toLowerCase().endsWith('.m3u8'));
                } catch {
                    return false;
                }
            });

        validateBtn.disabled = validUrls.length === 0;
        return validUrls;
    }

    // 复制URL
    function copyUrl(url) {
        navigator.clipboard.writeText(url).then(() => {
            const btn = document.querySelector('.btn-outline-success');
            const originalHtml = btn.innerHTML;
            
            // 更改图标为对勾
            btn.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-check" viewBox="0 0 16 16">
                    <path d="M10.97 4.97a.75.75 0 0 1 1.07 1.05l-3.99 4.99a.75.75 0 0 1-1.08.02L4.324 8.384a.75.75 0 1 1 1.06-1.06l2.094 2.093 3.473-4.425a.267.267 0 0 1 .02-.022z"/>
                </svg>
            `;

            // 显示 Toast 提示
            const toast = document.createElement('div');
            toast.className = 'toast-container position-fixed bottom-0 end-0 p-3';
            toast.innerHTML = `
                <div class="toast show" role="alert" aria-live="assertive" aria-atomic="true">
                    <div class="toast-header">
                        <svg class="bd-placeholder-img rounded me-2" width="20" height="20" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMidYMid slice" focusable="false">
                            <rect width="100%" height="100%" fill="#198754"></rect>
                        </svg>
                        <strong class="me-auto">提示</strong>
                        <button type="button" class="btn-close" data-bs-dismiss="toast" aria-label="Close"></button>
                    </div>
                    <div class="toast-body">
                        链接已复制到剪贴板
                    </div>
                </div>
            `;
            document.body.appendChild(toast);
            
            // 1秒后恢复原始图标并移除 Toast
            setTimeout(() => {
                btn.innerHTML = originalHtml;
                toast.remove();
            }, 1000);
        }).catch(err => {
            console.error('复制失败:', err);
            // 复制失败时显示错误 Toast
            const toast = document.createElement('div');
            toast.className = 'toast-container position-fixed bottom-0 end-0 p-3';
            toast.innerHTML = `
                <div class="toast show" role="alert" aria-live="assertive" aria-atomic="true">
                    <div class="toast-header">
                        <svg class="bd-placeholder-img rounded me-2" width="20" height="20" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMidYMid slice" focusable="false">
                            <rect width="100%" height="100%" fill="#dc3545"></rect>
                        </svg>
                        <strong class="me-auto">错误</strong>
                        <button type="button" class="btn-close" data-bs-dismiss="toast" aria-label="Close"></button>
                    </div>
                    <div class="toast-body">
                        复制失败，请手动复制链接
                    </div>
                </div>
            `;
            document.body.appendChild(toast);
            
            // 3秒后移除错误提示
            setTimeout(() => {
                toast.remove();
            }, 3000);
        });
    }

    // 更新进度文本
    function updateProgressText(progress, totalUrls) {
        const progressText = document.getElementById('progressText');
        if (progress < 30) {
            progressText.textContent = `正在获取 ${totalUrls} 个 M3U 文件...`;
        } else if (progress < 60) {
            progressText.textContent = '正在验证链接...';
        } else if (progress < 90) {
            progressText.textContent = '正在合并并生成缓存文件...';
        }
    }

    // 验证并提交
    function validateM3U() {
        const validUrls = validateAllInputs();
        const uploadToken = document.getElementById('uploadLabel').dataset.token;

        if (validUrls.length === 0 && !uploadToken) return;

        const validateBtn = document.getElementById('validateBtn');
        const validateBtnText = document.getElementById('validateBtnText');
        const validateSpinner = document.getElementById('validateSpinner');
        const progressArea = document.getElementById('progressArea');
        const progressBar = document.getElementById('progressBar');
        const progressText = document.getElementById('progressText');

        // 清空之前的结果
        document.getElementById('result').innerHTML = '';

        // 显示进度区域
        progressArea.classList.remove('d-none');
        progressBar.style.width = '0%';
        progressText.textContent = '正在处理...';
        
        // 禁用按钮并显示加载动画
        validateBtn.disabled = true;
        validateBtnText.textContent = '验证中...';
        validateSpinner.classList.remove('d-none');

        // 获取延迟设置
        const latency = parseInt(document.getElementById('latencyRange').value);

        // 发送请求到后端
        fetch('/api/validate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                urls: validUrls,
                maxLatency: latency,
                token: uploadToken || ''
            })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                const message = `${data.message}
原始链接：${data.stats.total} 个
验证通过：${data.stats.unique} 个
有效链接：${data.stats.valid} 个（去重后）

您可以通过以下地址访问合并后的 M3U 文件：
<a href="${data.m3uLink}" target="_blank">${data.m3uLink}</a>`;

                showResult('success', message.replace(/\n/g, '<br>'));
            } else {
                showResult('error', data.message);
            }
            validateBtn.disabled = false;
            validateBtnText.classList.remove('d-none');
            validateSpinner.classList.add('d-none');
            progressArea.classList.add('d-none');
        })
        .catch(error => {
            showResult('error', '验证请求失败，请稍后重试');
            validateBtn.disabled = false;
            validateBtnText.classList.remove('d-none');
            validateSpinner.classList.add('d-none');
            progressArea.classList.add('d-none');
        });
    }

    // 事件监听
    addUrlBtn.addEventListener('click', addUrlInput);
    validateBtn.addEventListener('click', validateM3U);

    // 初始化第一个输入框，但不显示删除按钮
    const firstInput = urlInputs.querySelector('.url-input-group');
    if (firstInput) {
        // 隐藏第一个输入框的删除按钮
        firstInput.querySelector('.btn-remove').style.display = 'none';
        // 为第一个输入框添加验证事件
        firstInput.querySelector('input').addEventListener('input', validateInput);
    }

    // 添加滑块事件监听
    const latencyRange = document.getElementById('latencyRange');
    latencyRange.addEventListener('input', updateLatencyValue);

    // 添加文件上传处理
    const fileUpload = document.getElementById('fileUpload');
    fileUpload.addEventListener('change', handleFileUpload);
});

// 添加滑块更新函数
function updateLatencyValue() {
    const value = document.getElementById('latencyRange').value;
    document.getElementById('latencyValue').textContent = value;
}

// 处理文件上传
function handleFileUpload(event) {
    const file = event.target.files[0];
    if (!file) return;

    // 检查文件类型
    if (!file.name.toLowerCase().endsWith('.m3u') && !file.name.toLowerCase().endsWith('.m3u8')) {
        showResult(`
            <div class="alert alert-danger" role="alert">
                <h4 class="alert-heading">错误</h4>
                <p>请上传 .m3u 或 .m3u8 格式的文件</p>
            </div>
        `);
        return;
    }

    const formData = new FormData();
    formData.append('file', file);
    formData.append('maxLatency', document.getElementById('latencyRange').value);

    // 更新上传按钮状态
    const uploadLabel = document.getElementById('uploadLabel');
    const uploadText = document.getElementById('uploadText');
    const uploadProgress = document.getElementById('uploadProgress');
    const progressBar = uploadProgress.querySelector('.progress-bar');
    const validateBtn = document.getElementById('validateBtn');
    
    // 更新按钮文本为处理状态
    uploadText.innerHTML = `
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-file-earmark-text me-1" viewBox="0 0 16 16">
            <path d="M5.5 7a.5.5 0 0 0 0 1h5a.5.5 0 0 0 0-1h-5zM5 9.5a.5.5 0 0 1 .5-.5h5a.5.5 0 0 1 0 1h-5a.5.5 0 0 1-.5-.5zm0 2a.5.5 0 0 1 .5-.5h2a.5.5 0 0 1 0 1h-2a.5.5 0 0 1-.5-.5z"/>
            <path d="M9.5 0H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V4.5L9.5 0zm0 1v2A1.5 1.5 0 0 0 11 4.5h2V14a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1h5.5z"/>
        </svg>
        正在处理: ${file.name}
    `;
    uploadProgress.classList.remove('d-none');
    uploadLabel.classList.add('disabled');

    fetch('/api/upload', {
        method: 'POST',
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // 更新上传按钮显示已上传的文件名
            uploadText.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-file-earmark-check me-1" viewBox="0 0 16 16">
                    <path d="M10.854 7.854a.5.5 0 0 0-.708-.708L7.5 9.793 6.354 8.646a.5.5 0 1 0-.708.708l1.5 1.5a.5.5 0 0 0 .708 0l3-3z"/>
                    <path d="M14 14V4.5L9.5 0H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2zM9.5 3A1.5 1.5 0 0 0 11 4.5h2V14a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1h5.5v2z"/>
                </svg>
                ${data.fileName}
            `;
            uploadLabel.classList.remove('btn-outline-secondary');
            uploadLabel.classList.add('btn-outline-success');
            uploadLabel.dataset.token = data.token;
            
            // 启用验证按钮
            validateBtn.disabled = false;
            
            showResult(`
                <div class="alert alert-success" role="alert">
                    <h4 class="alert-heading">文件上传成功！</h4>
                    <p>文件已成功上传，可以点击验证按钮开始处理。</p>
                </div>
            `);
        } else {
            throw new Error(data.message);
        }
    })
    .catch(error => {
        // 恢复上传按钮状态
        uploadText.innerHTML = `
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-file-earmark-text me-1" viewBox="0 0 16 16">
                <path d="M5.5 7a.5.5 0 0 0 0 1h5a.5.5 0 0 0 0-1h-5zM5 9.5a.5.5 0 0 1 .5-.5h5a.5.5 0 0 1 0 1h-5a.5.5 0 0 1-.5-.5zm0 2a.5.5 0 0 1 .5-.5h2a.5.5 0 0 1 0 1h-2a.5.5 0 0 1-.5-.5z"/>
                <path d="M9.5 0H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V4.5L9.5 0zm0 1v2A1.5 1.5 0 0 0 11 4.5h2V14a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1h5.5z"/>
            </svg>
            上传本地文件
        `;
        uploadLabel.classList.remove('disabled', 'btn-outline-success');
        uploadLabel.classList.add('btn-outline-secondary');
        uploadProgress.classList.add('d-none');
        delete uploadLabel.dataset.token;
        
        showResult(`
            <div class="alert alert-danger" role="alert">
                <h4 class="alert-heading">处理失败</h4>
                <p>${error.message || '文件处理失败，请稍后重试'}</p>
            </div>
        `);
    })
    .finally(() => {
        // 清理文件输入但保留文件名显示
        event.target.value = '';
        // 隐藏进度条
        setTimeout(() => {
            uploadProgress.classList.add('d-none');
        }, 1000);
    });

    // 模拟上传进度
    let progress = 0;
    const progressInterval = setInterval(() => {
        if (progress < 90) {
            progress += 10;
            progressBar.style.width = progress + '%';
        }
    }, 500);

    // 在成功或失败时清除进度条计时器
    setTimeout(() => {
        clearInterval(progressInterval);
        progressBar.style.width = '100%';
    }, 3000);
}

// ... 保留之前的其他函数 ... 