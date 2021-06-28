$(document).ready(function () {
    //登录
    $("#login-form").validate({
        rules: {
            Name: {
                required: true,
                rangelength: [5, 10]
            },
            RoomId: {
                required: true,
                rangelength: [1, 5]
            }
        },
        messages: {
            Nme: {
                required: "请输入用户名",
                rangelength: "用户名必须是5-10位"
            },
            RoomId: {
                required: "请输入房间号",
                rangelength: "密码必须是1-5位"
            }
        },
        submitHandler: function (form) {
            var urlStr = "/login"
            $(form).ajaxSubmit({
                url: urlStr,
                type: "post",
                dataType: "json",
                success: function (data, status) {
                    alert("warning:" + data.msg + " status:" + status)
                    if (data.code == 301) {
                        setTimeout(function () {
                            window.location.href = "/chat_room?user_name=" + data.u_Name + "&room_id=" + data.r_Id
                        }, 500)
                    }
                },
                error: function (data, status) {
                    alert("err:" + data.message + ":" + status)
                }
            });
        }
    });

});
