document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('uploadForm');
    const fileInput = document.getElementById('excelFile');
    const fileChosen = document.getElementById('file-chosen');
    const submitButton = form.querySelector('button[type="submit"]');
    const pdfPreview = document.getElementById('pdfPreview');
    const pdfFrame = document.getElementById('pdfFrame');
    const downloadButton = document.getElementById('downloadPdf');

    let pdfBlob = null;


    fileInput.addEventListener('change', function() {
        if (this.files && this.files.length > 0) {
            fileChosen.textContent = this.files[0].name;
        } else {
            fileChosen.textContent = 'Файл не выбран';
        }
    });
    form.addEventListener('submit', function(e) {
        e.preventDefault();

        if (!fileInput.files || fileInput.files.length === 0) {
            alert('Пожалуйста, выберите файл Excel');
            return;
        }

        const formData = new FormData(form);

        submitButton.disabled = true;
        submitButton.textContent = 'Обработка...';

        fetch('/upload', {
            method: 'POST',
            body: formData
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Ошибка сервера');
            }
            return response.blob();
        })
        .then(blob => {
            pdfBlob = blob;
            const url = window.URL.createObjectURL(blob);
            
            // Показываем предварительный просмотр
            pdfFrame.src = url;
            pdfPreview.style.display = 'block';
            
            submitButton.textContent = 'PDF успешно создан!';
            setTimeout(() => {
                submitButton.textContent = 'Сгенерировать PDF';
                submitButton.disabled = false;
            }, 3000);
        })
        .catch(error => {
            console.error('Error:', error);
            submitButton.textContent = 'Ошибка. Попробуйте снова';
            submitButton.disabled = false;
        });
    });

});