function update() {
    let password = document.querySelector('#password_check').value;

    fetch("http://localhost/auth/sign-in/forgot-pass/", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({"password": password})
    })
        .then(res => {
            if (res.redirected) {
                document.location = res.url
            }
        })
}