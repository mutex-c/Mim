package com.mim.im.api

import com.google.gson.GsonBuilder
import com.mim.im.data.local.PreferencesHelper
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

class ApiClient(private val prefs: PreferencesHelper) {

    private var currentBaseUrl: String = ""
    private var cachedService: ApiService? = null
    private var cachedOkHttpClient: OkHttpClient? = null

    val okHttpClient: OkHttpClient
        get() {
            if (cachedOkHttpClient == null) {
                val logging = HttpLoggingInterceptor().apply {
                    level = HttpLoggingInterceptor.Level.BODY
                }

                val authInterceptor = Interceptor { chain ->
                    val original = chain.request()
                    val token = prefs.token
                    val builder = original.newBuilder()
                    if (!token.isNullOrEmpty()) {
                        builder.header("Authorization", "Bearer $token")
                    }
                    chain.proceed(builder.build())
                }

                cachedOkHttpClient = OkHttpClient.Builder()
                    .connectTimeout(15, TimeUnit.SECONDS)
                    .readTimeout(30, TimeUnit.SECONDS)
                    .writeTimeout(30, TimeUnit.SECONDS)
                    .addInterceptor(authInterceptor)
                    .addInterceptor(logging)
                    .build()
            }
            return cachedOkHttpClient!!
        }

    val service: ApiService
        get() {
            val url = prefs.serverUrl.trimEnd('/') + "/"
            if (cachedService != null && currentBaseUrl == url) {
                return cachedService!!
            }

            val gson = GsonBuilder()
                .setLenient()
                .create()

            val retrofit = Retrofit.Builder()
                .baseUrl(url)
                .client(okHttpClient)
                .addConverterFactory(GsonConverterFactory.create(gson))
                .build()

            currentBaseUrl = url
            cachedService = retrofit.create(ApiService::class.java)
            return cachedService!!
        }

    fun invalidate() {
        cachedService = null
        cachedOkHttpClient = null
    }
}
