document.addEventListener('DOMContentLoaded', function() {
    const urlInputs = document.getElementById('urlInputs');
    const addUrlBtn = document.getElementById('addUrlBtn');
    const validateBtn = document.getElementById('validateBtn');
    const template = document.getElementById('urlInputTemplate');

    // 添加链接输入框
    function addUrlInput() {
        const clone = template.content.cloneNode(true);
        urlInputs.appendChild(clone);
        
        const newInput = urlInputs.lastElementChild;
        const input = newInput.querySelector('input');
        const removeBtn = newInput.querySelector('.btn-remove');

        input.addEventListener('input', validateInput);
        removeBtn.addEventListener('click', () => removeUrlInput(newInput));

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
            return isValid;
        } catch {
            inputGroup.classList.toggle('is-invalid', url !== '');
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

    // 验证并提交
    function validateM3U() {
        const validUrls = validateAllInputs();
        if (validUrls.length === 0) return;

        const validateBtn = document.getElementById('validateBtn');
        const validateBtnText = document.getElementById('validateBtnText');
        const validateSpinner = document.getElementById('validateSpinner');
        const progressArea = document.getElementById('progressArea');
        const progressBar = document.getElementById('progressBar');
        const progressText = document.getElementById('progressText');

        // 显示进度区域
        progressArea.classList.remove('d-none');
        progressBar.style.width = '0%';
        progressText.textContent = `正在处理 ${validUrls.length} 个链接...`;
        
        // 禁用按钮并显示加载动画
        validateBtn.disabled = true;
        validateBtnText.textContent = '验证中...';
        validateSpinner.classList.remove('d-none');

        // ... 其余验证逻辑保持不变 ...
    }

    // 事件监听
    addUrlBtn.addEventListener('click', addUrlInput);
    validateBtn.addEventListener('click', validateM3U);

    // 初始化第一个输入框
    addUrlInput();
});

// ... 保留之前的其他函数 ... 