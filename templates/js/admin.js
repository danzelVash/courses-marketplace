function sign_in_admin() {
    let login = document.querySelector('#admin_login');
    let password = document.querySelector('#admin_password');

    fetch("http://localhost/admin/auth/", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({"admin_login": login.value, "admin_password": password.value})
    })
        .then(res => {
            if (res.redirected) {
                document.location = res.url
            }
        })
}