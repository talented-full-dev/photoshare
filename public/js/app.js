(function() {
    'use strict';
    // Declare app level module which depends on filters, and services
    angular.module('photoshare', [
        'ngRoute',
        'ngResource',
        'ngAnimate',
        'ngSanitize',
        'ngCookies',
        'ui.gravatar',
        'photoshare.filters',
        'photoshare.services',
        'photoshare.directives',
        'photoshare.controllers'
    ]).
    constant('urls', {
        auth: '/api/auth/',
        photos: '/api/photos/:id',
        tags: '/api/tags/',
        messages: '/api/messages'
    }).
    constant('authTokenHeader', 'X-Auth-Token').
    constant('authTokenStorageKey', 'authToken').
    config(['$routeProvider',
        '$locationProvider',
        '$httpProvider',
        '$resourceProvider',
        function(
            $routeProvider,
            $locationProvider,
            $httpProvider,
            $resourceProvider
        ) {
            $routeProvider.

            when('/popular', {
                templateUrl: 'partials/list.html',
                controller: 'ListCtrl'
            }).

            when('/latest', {
                templateUrl: 'partials/list.html',
                controller: 'ListCtrl'
            }).

            when('/tags', {
                templateUrl: 'partials/tags.html',
                controller: 'TagsCtrl'
            }).

            when('/tag/:tag', {
                templateUrl: 'partials/list.html',
                controller: 'ListCtrl'
            }).

            when('/search/:q', {
                templateUrl: 'partials/list.html',
                controller: 'ListCtrl'
            }).

            when('/owner/:ownerID/:ownerName', {
                templateUrl: 'partials/list.html',
                controller: 'ListCtrl'
            }).

            when('/detail/:id', {
                templateUrl: 'partials/detail.html',
                controller: 'DetailCtrl'
            }).

            when('/upload', {
                templateUrl: 'partials/upload.html',
                controller: 'UploadCtrl',
                loginRequired: true
            }).

            when('/login', {
                templateUrl: 'partials/login.html',
                controller: 'LoginCtrl'
            }).

            when('/recoverpass', {
                templateUrl: 'partials/recover_pass.html',
                controller: 'RecoverPassCtrl'
            }).

            when('/changepass', {
                templateUrl: 'partials/change_pass.html',
                controller: 'ChangePassCtrl'
            }).

            when('/signup', {
                templateUrl: 'partials/signup.html',
                controller: 'SignupCtrl'
            }).

            otherwise({
                redirectTo: '/popular'
            });
            //$locationProvider.html5Mode(true);
            //
            $resourceProvider.defaults.stripTrailingSlashes = false;

            //$httpProvider.defaults.xsrfCookieName = "csrf_token";
            //$httpProvider.defaults.xsrfHeaderName = "X-CSRF-Token";

            // handle file uploads

            $httpProvider.defaults.transformRequest = function(data, headersGetter) {

                if (data === undefined) {
                    return data;
                }

                var fd = new FormData(),
                    isFileUpload = false,
                    headers = headersGetter();

                angular.forEach(data, function(value, key) {
                    if (value instanceof FileList) {
                        isFileUpload = true;
                        if (value.length === 1) {
                            fd.append(key, value[0]);
                        } else {
                            angular.forEach(value, function(file, index) {
                                fd.append(key + "_" + index, file);
                            });
                        }
                    } else {
                        fd.append(key, value);
                    }
                });
                if (isFileUpload) {
                    headers["Content-Type"] = undefined;
                    return fd;
                }

                return JSON.stringify(data);
            };

            var interceptors = ['AuthInterceptor', 'ErrorInterceptor'];

            angular.forEach(interceptors, function(interceptor) {
                $httpProvider.interceptors.push([
                    '$injector',
                    function($injector) {
                        return $injector.get(interceptor);
                    }
                ]);
            });

        }
    ]).factory('AuthInterceptor', function($window, $cookies, authTokenHeader, authTokenStorageKey) {

        return {
            request: function(config) {

                // oauth2 authentication sets token in cookie before redirect
                // check the existence of cookie, add to local storage, and delete the cookie.
                if ($cookies.authToken) {
                    $window.localStorage.setItem(authTokenStorageKey, $cookies.authToken);
                    delete $cookies.authToken;
                }

                config.headers = config.headers || {};

                var token = $window.localStorage.getItem(authTokenStorageKey);

                if (token) {
                    config.headers[authTokenHeader] = token;
                }
                return config;
            }
        };

    }).factory('ErrorInterceptor', function($q, $location, Session, Alert) {
        return {

            response: function(response) {
                return response;
            },

            responseError: function(response) {
                var rejection = $q.reject(response),
                    status = response.status,
                    msg = 'Sorry, an error has occurred';

                if (status == 401) {
                    Session.redirectToLogin();
                    return;
                }
                if (status == 404) {
                    // handle locally
                    return;
                }
                if (status == 403) {
                    msg = "Sorry, you're not allowed to do this";
                }
                if (status == 400 && response.data.errors) {
                    msg = "Sorry, your form contains errors, please try again";
                }
                if (status == 413) {
                    msg = "The file was too large!";
                }
                if (response.data && typeof(response.data) === 'string') {
                    msg = response.data;
                }
                if (msg) {
                    Alert.danger(msg);
                }
                return rejection;
            }
        };
    });
})();
