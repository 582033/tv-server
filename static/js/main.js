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
    const uploadLabel = document.getElementById('uploadLabel');

    // 添加链接输入框
    function addUrlInput(value = '') {
        const template = document.getElementById('urlInputTemplate');
        if (!template) {
            console.error('找不到输入框模板');
            return;
        }

        const newInput = template.content.cloneNode(true);
        const inputGroup = newInput.querySelector('.url-input-group');
        const input = newInput.querySelector('input');
        const removeBtn = newInput.querySelector('.btn-remove');

        if (input) {
            input.addEventListener('input', validateInput);
            if (value) {
                input.value = value;
            }
        }

        if (removeBtn) {
            // removeBtn.addEventListener('click', () => {
            //     if (inputGroup) {
            //         removeUrlInput(inputGroup);
            //     }
            // });
        }

        urlInputs.appendChild(newInput);
        validateAllInputs();
        return inputGroup;
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
有效链接：${data.stats.valid} 个

您可以通过以下地址访问合并后的 M3U 文件：
<a href="${data.m3uLink}" target="_blank">${data.m3uLink}</a>`;

                showResult('success', message.replace(/\n/g, '<br>'));
            } else {
                showResult('error', data.message);
            }
            validateBtn.disabled = false;
            validateBtnText.textContent = '验证';  
            validateBtnText.classList.remove('d-none');
            validateSpinner.classList.add('d-none');
            progressArea.classList.add('d-none');
        })
        .catch(error => {
            showResult('error', '验证请求失败，请稍后重试');
            validateBtn.disabled = false;
            validateBtnText.textContent = '验证';  
            validateBtnText.classList.remove('d-none');
            validateSpinner.classList.add('d-none');
            progressArea.classList.add('d-none');
        });
    }

    // 事件监听
    addUrlBtn.addEventListener('click', () => addUrlInput());
    validateBtn.addEventListener('click', validateM3U);

    // 使用事件委托处理所有删除按钮的点击
    document.addEventListener('click', function(e) {
        const deleteBtn = e.target.closest('.btn-remove');
        if (deleteBtn) {
            e.preventDefault();
            e.stopPropagation();
            
            const inputGroup = deleteBtn.closest('.url-input-group');
            if (inputGroup) {
                removeUrlInput(inputGroup);
            }
        }

        const uploadDeleteBtn = e.target.closest('.upload-remove-btn');
        if (uploadDeleteBtn && uploadDeleteBtn.parentElement === uploadLabel) {
            e.preventDefault();
            e.stopPropagation();
            
            const uploadText = document.getElementById('uploadText');
            
            // 恢复上传按钮状态
            uploadText.innerHTML = '上传本地文件';
            uploadLabel.className = 'btn btn-outline-secondary';
            uploadLabel.removeAttribute('data-token');
            
            // 清除文件输入框的值
            document.getElementById('fileUpload').value = '';
            
            // 禁用验证按钮
            document.getElementById('validateBtn').disabled = true;
            
            // 清除结果显示
            document.getElementById('result').innerHTML = '';
            
            // 隐藏进度条
            document.getElementById('uploadProgress').classList.add('d-none');
            
            // 移除删除按钮
            uploadDeleteBtn.remove();
        }
    });    

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
    
    // 更新按钮文本，不显示文件图标
    uploadText.innerHTML = `正在处理: ${file.name}`;
    uploadProgress.classList.remove('d-none');
    uploadLabel.classList.add('disabled');

    fetch('/api/upload', {
        method: 'POST',
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // 更新按钮文本，不显示文件图标
            uploadText.innerHTML = `已上传: ${file.name}`;
            uploadLabel.classList.add('btn-outline-success', 'position-relative');
            uploadLabel.dataset.token = data.token;

            // 创建一个包装容器
            const wrapper = document.createElement('div');
            wrapper.className = 'position-relative d-inline-block w-100';
            uploadLabel.parentNode.insertBefore(wrapper, uploadLabel);
            wrapper.appendChild(uploadLabel);

            // 添加删除按钮
            const deleteIcon = document.createElement('button');
            deleteIcon.type = 'button';
            deleteIcon.className = 'badge rounded-circle border-0 bg-danger position-absolute top-0 end-0';
            deleteIcon.innerHTML = '×';
            deleteIcon.style.cursor = 'pointer';
            deleteIcon.style.transform = 'translate(50%, -50%)';
            deleteIcon.setAttribute('aria-label', '删除');
            
            // 添加点击事件
            deleteIcon.onclick = function(e) {
                e.preventDefault();
                e.stopPropagation();
                
                // 恢复上传按钮状态
                uploadText.innerHTML = '上传本地文件';
                uploadLabel.className = 'btn btn-outline-secondary btn-sm w-100';
                uploadLabel.removeAttribute('data-token');
                
                // 清除文件输入框的值
                document.getElementById('fileUpload').value = '';
                
                // 禁用验证按钮
                document.getElementById('validateBtn').disabled = true;
                
                // 清除结果显示
                document.getElementById('result').innerHTML = '';
                
                // 隐藏进度条
                document.getElementById('uploadProgress').classList.add('d-none');
                
                // 移除包装容器（包括删除按钮）
                wrapper.parentNode.insertBefore(uploadLabel, wrapper);
                wrapper.remove();
            };
            
            wrapper.appendChild(deleteIcon);
            
            // 启用验证按钮
            validateBtn.disabled = false;
        } else {
            throw new Error(data.message || '上传失败');
        }
    })
    .catch(error => {
        // 恢复上传按钮状态
        uploadText.innerHTML = `
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-cloud-upload me-1" viewBox="0 0 16 16">
                <path fill-rule="evenodd" d="M4.406 1.342A5.53 5.53 0 0 1 8 0c2.69 0 4.923 2 5.166 4.579C14.758 4.804 16 6.137 16 7.773 16 9.569 14.502 11 12.687 11H10a.5.5 0 0 1 0-1h2.688C13.979 10 15 8.988 15 7.773c0-1.216-1.02-2.228-2.313-2.228h-.5v-.5C12.188 2.825 10.328 1 8 1a4.53 4.53 0 0 0-2.941 1.1c-.757.652-1.153 1.438-1.153 2.055v.448l-.445.049C2.064 4.805 1 5.952 1 7.318 1 8.785 2.23 10 3.781 10H6a.5.5 0 0 1 0 1H3.781C1.708 11 0 9.366 0 7.318c0-1.763 1.266-3.223 2.942-3.593.143-.863.698-1.723 1.464-2.383z"/>
                <path fill-rule="evenodd" d="M7.646 4.146a.5.5 0 0 1 .708 0l3 3a.5.5 0 0 1-.708.708L8.5 5.707V14.5a.5.5 0 0 1-1 0V5.707L5.354 7.854a.5.5 0 1 1-.708-.708l3-3z"/>
            </svg>
            上传本地文件
        `;
        uploadLabel.classList.remove('disabled', 'btn-outline-success');
        uploadLabel.classList.add('btn-outline-secondary');
        delete uploadLabel.dataset.token;
        uploadProgress.classList.add('d-none');
        
        // 显示错误信息
        // showResult('error', error.message || '上传失败，请稍后重试');
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

// 添加清除上传的函数
function clearUpload(event) {
    if (event) {
        event.preventDefault();
        event.stopPropagation();
    }
    
    const uploadLabel = document.getElementById('uploadLabel');
    const uploadText = document.getElementById('uploadText');
    const fileUpload = document.getElementById('fileUpload');
    
    // 恢复上传按钮状态
    uploadText.innerHTML = `
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-cloud-upload me-1" viewBox="0 0 16 16">
            <path fill-rule="evenodd" d="M4.406 1.342A5.53 5.53 0 0 1 8 0c2.69 0 4.923 2 5.166 4.579C14.758 4.804 16 6.137 16 7.773 16 9.569 14.502 11 12.687 11H10a.5.5 0 0 1 0-1h2.688C13.979 10 15 8.988 15 7.773c0-1.216-1.02-2.228-2.313-2.228h-.5v-.5C12.188 2.825 10.328 1 8 1a4.53 4.53 0 0 0-2.941 1.1c-.757.652-1.153 1.438-1.153 2.055v.448l-.445.049C2.064 4.805 1 5.952 1 7.318 1 8.785 2.23 10 3.781 10H6a.5.5 0 0 1 0 1H3.781C1.708 11 0 9.366 0 7.318c0-1.763 1.266-3.223 2.942-3.593.143-.863.698-1.723 1.464-2.383z"/>
            <path fill-rule="evenodd" d="M7.646 4.146a.5.5 0 0 1 .708 0l3 3a.5.5 0 0 1-.708.708L8.5 5.707V14.5a.5.5 0 0 1-1 0V5.707L5.354 7.854a.5.5 0 1 1-.708-.708l3-3z"/>
        </svg>
        上传本地文件
    `;
    uploadLabel.classList.remove('disabled', 'btn-outline-success');
    uploadLabel.classList.add('btn-outline-secondary');
    delete uploadLabel.dataset.token;
    uploadProgress.classList.add('d-none');
    
    // 清除文件输入框的值
    if (fileUpload) {
        fileUpload.value = '';
    }
    
    // 禁用验证按钮
    const validateBtn = document.getElementById('validateBtn');
    if (validateBtn) {
        validateBtn.disabled = true;
    }
    
    // 清除结果显示
    const result = document.getElementById('result');
    if (result) {
        result.innerHTML = '';
    }
    
    // 隐藏进度条
    const uploadProgress = document.getElementById('uploadProgress');
    if (uploadProgress) {
        uploadProgress.classList.add('d-none');
    }
}
