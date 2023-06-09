function check_admin_auth_code() {
    let code = document.querySelector('#auth_code');

    fetch("http://localhost/admin/auth/code/", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({"code": code.value})
    })
        .then(res => {
            if (res.redirected) {
                document.location = res.url
            }
        })
}