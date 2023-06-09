function sign_out() {
    fetch("http://localhost/auth/sign-out/", {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(res => {
            if (res.redirected) {
                document.location = res.url
            }
        });
}