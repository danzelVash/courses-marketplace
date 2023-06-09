const login = document.querySelector('.login');
const password = document.querySelector('.password');
const textError = document.querySelector('.textError');


function checkLoginFormat(value) {  // проверка на верный формат логина
    let emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(value)) {
        incorrectFormatLogin();
        return false;
    } else {
        return true;
    }
}

function checkPasswordFormat(value) {
    let format = value !== "";
    if (format === false && !password.classList.contains('redOutline')) {
        incorrectFormatPassword();
    }
    return format;
}


function incorrectFormatLogin() {  // ф-ия добавления красных полей и текста (для логина)
    addRedOutline(login, password);
    textError.classList.add('active');
    textError.innerText = "*Неверный формат логина";
}

function incorrectFormatPassword() {  // ф-ия добавления красных полей и текста (для пароля)
    addRedOutline(password);
    textError.classList.add('active');
    textError.innerText = "*Пожалуйста, введите пароль";
}


function addRedOutline() {
    for (let value of arguments) {
        value.classList.add("redOutline");
    }
}

function removeRedOutline() {
    for (let value of arguments) {
        value.classList.remove("redOutline");
    }
}


function clearText() {
    textError.innerText = "";
    textError.classList.remove('active');
}


function focusL(target) { // при нажатии на поле логина
    if (login.classList.contains('redOutline')) {  // если имеется класс - удалить его
        removeRedOutline(login, password);
        clearText();
    }
}

function focusP(target) { // при нажатии на поле пароля
    if (!login.classList.contains('redOutline') && password.classList.contains('redOutline')) {
        removeRedOutline(password);
        clearText();
    }
}


function Error500() {
    addRedOutline(password);
    textError.classList.add('active');
    textError.innerText = "*Неожиданная ошибка, пожалуйста повторите попытку позже!";
}

function Error400(json) {
    addRedOutline(login, password);
    textError.classList.add('active');
    console.log(json)
    if (json.exist && !json.right_pass) {
        textError.innerText = "*Неверный пароль, попробуйте еще раз!";
    } else if (!json.exist) {
        textError.innerText = "*Аккаунт с таким логином не зарегистрирован на сайте!";
    }

}


document.addEventListener("keyup", function (event) {
    if (event.keyCode === 13) {
        event.preventDefault();
        sign_in();
    }
});


/* Любое изменение приравнивается к фокусу на поля */

login.oninput = function () {
    focusL();
}
password.oninput = function () {
    focusP();
}

/**/

function sign_in() {


    let key1 = checkLoginFormat(login.value);
    let key2 = false;
    if (key1 === true) {
        key2 = checkPasswordFormat(password.value);
    }


    if (key1 && key2) {

        fetch("http://localhost/auth/sign-in/", {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({"login": login.value, "password": password.value})
        })
            .then(res => {
                if (res.redirected) {
                    document.location = res.url;
                } else if (res.status === 500) {
                    return res.json()
                        .then(data => {
                            Error500(data)
                        })
                } else if (res.status === 200) {
                    console.log("email is exist and code was sent")
                } else if (res.status === 400) {
                    return res.json()
                        .then(data => {
                            Error400(data);
                        })
                }
            })
    }
}

