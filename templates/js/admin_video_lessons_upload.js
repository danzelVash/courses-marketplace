// function add_lesson() {
//     let title = document.querySelector('#title').value;
//     let description = document.querySelector('#description').value;
//     let video = document.querySelector('#video').files;
//     let video_name = document.querySelector('#video_name').value;
//     let files = document.querySelector('#additional_files').files;
//     let additional_filenames = document.querySelector('#additional_filenames').value.split(", ");
//     let price = document.querySelector('#price').value;
//
//     console.log(additional_filenames)
//     let masProm = [];
//     for (let i = 0; i < files.length; i++) {
//         masProm.push(new Promise(resolve => {
//             let reader = new FileReader();
//             reader.nameFile = additional_filenames[i];
//             reader.onload = () => resolve(reader);
//             reader.readAsDataURL(files[i])
//         }))
//     }
//     Promise.all(masProm)
//         .then(binFiles => {
//             let additional_files = [];
//             let i = 0
//             binFiles.forEach(binFile => {
//                 additional_files.push({
//                     "data": binFile.result.split(",")[1],
//                     "file_name": additional_filenames[i],
//                 })
//                 i++
//             })
//             return additional_files;
//         })
//         .then(result => {
//             let videoProm = [];
//             videoProm.push(new Promise(resolve => {
//                 let reader = new FileReader();
//                 reader.nameFile = video_name;
//                 reader.onload = () => resolve(reader);
//                 reader.readAsDataURL(video[0])
//             }))
//             Promise.all(videoProm)
//                 .then(binFiles => {
//                     let videoData = ""
//                     binFiles.forEach(binFile => {
//                         videoData = binFile.result.split(",")[1]
//                     })
//                     return videoData
//                 })
//                 .then(res => {
//                     let body = {
//                         "title": title,
//                         "description": description,
//                         "video": res,
//                         "video_name": video_name,
//                         "additional_files": result,
//                         "price": price,
//
//                     }
//                     console.log(body)
//                     fetch("http://localhost:8080/admin/panel/video-lessons/upload/", {
//                         method: 'POST',
//                         headers: {
//                             'Content-Type': 'application/json'
//                         },
//                         body: JSON.stringify(body)
//                     })
//                         .then(response => {
//                             //     обработку ответа надо сделать
//                         })
//                 })
//
//         })
//
//
// }

function add_lesson() {
    const title = document.querySelector('#title').value;
    const description = document.querySelector('#description').value;
    const video = document.querySelector('#video').files[0];
    const video_name = document.querySelector('#video_name').value;
    const additional_files = document.querySelector('#additional_files').files;
    const additional_filenames = document.querySelector('#additional_filenames').value.split(",");
    const price = document.querySelector('#price').value;

    const formData = new FormData();

    formData.append('title', title);
    formData.append('description', description);
    formData.append('price', price);
    formData.append('video', video);
    formData.append('video_name', video_name);

    for (let i = 0; i < additional_files.length; i++) {
        formData.append('additional_files', additional_files[i]);
        formData.append('additional_filenames', additional_filenames[i])
    }

    const xhr = new XMLHttpRequest();

    xhr.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
            console.log(this.responseText);
        }
    };

    xhr.open('POST', 'http://localhost/admin/panel/video-lessons/upload/');
    xhr.send(formData);

}

// const form = document.querySelector('.form-data');
// form.addEventListener('submit', function (e) {
//     e.preventDefault();
//     let elem = e.target
//
//     let formData = new FormData();
//     formData.append('title', elem.querySelector('[name="title"]').value);
//     formData.append('description', elem.querySelector('[name="description"]').value);
//     formData.append('video', elem.querySelector('[name="video"]').files[0]);
//     formData.append('video_name', elem.querySelector('[name="video_name"]').value);
//     formData.append('additional_files', elem.querySelector('[name="additional_files"]').files);
//     formData.append('additional_filenames', elem.querySelector('[name="additional_filenames"]').value.split(","));
//     formData.append('price', elem.querySelector('[name="price"]').value);
//
//     axios.post("http://localhost:8080/admin/panel/video-lessons/upload/", formData)
//         .then(res => console.log(res))
//         .catch(err => console.log(err))
// })
