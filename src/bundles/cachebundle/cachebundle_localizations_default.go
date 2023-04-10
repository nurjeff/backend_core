package cachebundle

var valid_locales map[string]bool = map[string]bool{}

var default_locales map[string]map[string]string = map[string]map[string]string{
	"en_EN": default_en,
	"de_DE": default_de,
}

var default_en map[string]string = map[string]string{
	"bad_login":                        "Bad login",
	"not_found":                        "Not found",
	"auth_error":                       "Authentification error",
	"not_authorized":                   "Not authorized",
	"internal_error":                   "Internal error",
	"male":                             "Male",
	"female":                           "Female",
	"diverse":                          "Diverse",
	"err_username_empty":               "Username can't be empty",
	"err_username_illegal_characters":  "Username contains illegal characters",
	"err_password_empty":               "Password can't be empty",
	"err_password_no_match":            "Passwords don't match",
	"err_email_invalid":                "E-mail is not valid",
	"err_account_exists":               "Account with these credentials already exists",
	"err_telephone_exists":             "Account with this number already exists",
	"err_telephone_invalid":            "Mobile number is invalid",
	"err_gender_empty":                 "Gender can't be empty",
	"err_country_empty":                "Country can't be empty",
	"err_country_telephone_dont_match": "Country does not match phone number",
	"err_not_authorized":               "You are not authorized to access this",
	"suffix_wish_shared":               "shared a new wish with you!",
	"err_fcm_empty":                    "FCM-key is empty",
	"message":                          "Message",
	"successfully_delivered":           "sucessfully delivered",
	"error_delivered":                  "could not be delivered",
}

var default_de map[string]string = map[string]string{
	"bad_login":                        "Login ungültig",
	"not_found":                        "Nicht gefunden",
	"auth_error":                       "Authentifizierungsfehler",
	"not_authorized":                   "Nicht autorisiert",
	"internal_error":                   "Interner Fehler",
	"male":                             "Männlich",
	"female":                           "Weiblich",
	"diverse":                          "Divers",
	"err_username_empty":               "Benutzername kann nicht leer sein",
	"err_username_illegal_characters":  "Benutzername enthält unerlaubte Zeichen",
	"err_password_empty":               "Passwort kann nicht leer sein",
	"err_password_no_match":            "Passwörter stimmen nicht überein",
	"err_email_invalid":                "E-Mail ist nicht gültig",
	"err_account_exists":               "Ein Account mit diesen Daten existiert bereits",
	"err_telephone_exists":             "Ein Account mit dieser Telefonnummer existiert bereits",
	"err_telephone_invalid":            "Telefonnummer ist ungültig",
	"err_gender_empty":                 "Geschlecht kann nicht leer sein",
	"err_country_empty":                "Land kann nicht leer sein",
	"err_country_telephone_dont_match": "Land und Telefonnummer stimmen nicht überein",
	"err_not_authorized":               "Du hast keine Berechtigung um Dies zu tun",
	"suffix_wish_shared":               "hat einen neuen Wunsch mit dir geteilt!",
	"err_fcm_empty":                    "FCM-Schlüssel kann nicht leer sein",
	"message":                          "Nachricht",
	"successfully_delivered":           "wurde erfolgreich verschickt",
	"error_delivered":                  "konnte nicht verschickt werden",
}
