<script lang="ts">
    import Box from "./Box.svelte";
    import { createSnackbar } from "../../global";
    import { onMount } from "svelte";

    const mainIcons = {
        Clear: ["light_mode", "nightlight"],
        Rain: ["cloudy_snowing", "cloudy_snowing"],
        Clouds: ["cloud", "nights_stay"],
        Snow: ["ac_unit", "ac_unit"],
        Thunderstorm: ["thunderstorm", "thunderstorm"],
        Drizzle: ["grain", "grain"],
    };

    let loading = false;
    let loaded = false;
    let cachedOnly = false;

    let isNight = true;

    let data: weatherData = {
        id: 0,
        time: 0,
        weatherTitle: "",
        weatherDescription: "unknown",
        temperature: 0,
        feelsLike: 0,
        humidity: 0,
        sunrise: 0,
        sunset: 0,
    };

    interface weatherData {
        id: number;
        time: number;
        weatherTitle: string;
        weatherDescription: string;
        temperature: number;
        feelsLike: number;
        humidity: number;
        sunrise: number;
        sunset: number;
    }

    // Is used in case the normal weather data is not fetchable due to broken network conditions
    // Requests a max. 30 minute old snapshot of the weather data from the server
    async function loadCachedWeatherData(): Promise<weatherData> {
        let res = await (await fetch("/api/weather/cached")).json();
        if (res.success !== undefined && !res.success) throw Error(res.error);
        return res;
    }

    // Fetches a current (max. 5 minutes old) weather from the server
    async function loadWeatherData() {
        loading = true;
        try {
            let res = await fetch("/api/weather");

            // If the request fails due to the server failing unexpectedly, try using the cached data instead
            if (res.status === 500) {
                data = await loadCachedWeatherData();
                cachedOnly = true;
                $createSnackbar(
                    `Warning: Using fallback weather data from cache due to server error`
                );
                return;
            }

            const resTemp = await res.json();
            if (resTemp.success !== undefined && !resTemp.success)
                throw Error(resTemp.error);

            // Signal that the weather has been loaded successfully
            data = resTemp;

            // Detect if the night icon should be used
            const sunrise = new Date(data.sunrise)
            const sunset = new Date(data.sunset)
            const now = new Date()

            // Actual detection
            isNight = now.getHours() >= sunset.getHours() || now.getHours() < sunrise.getHours()

            loaded = true;
        } catch (err) {
            $createSnackbar(`Could not load weather: ${err}`);
        }
        loading = false;
    }

    onMount(loadWeatherData);
</script>

<Box bind:loading>
    <span slot="header">Weather</span>
    <div class="weather" slot="content">
        <div class="weather__top">
            {#if loaded}
                <i class="material-icons weather__top__icon">
                    {isNight
                        ? mainIcons[data.weatherTitle][1]
                        : mainIcons[data.weatherTitle][0]}
                </i>
            {/if}
            <div class="weather__top__labels">
                <div class="weather__top__labels__title">
                    {data.weatherTitle}
                </div>
                <div class="weather__top__labels__description">
                    {data.weatherDescription}
                </div>
            </div>
        </div>
        <div class="weather__measurements">
            <div class="weather__measurements__measurement">
                <div class="weather__measurements__temperature">
                    {Math.round(data.temperature)}°C
                </div>
                <span class="text-hint">Temperature</span>
            </div>
            <div class="weather__measurements__measurement">
                <div class="weather__measurements__feels-like">
                    {Math.round(data.feelsLike)}°C
                </div>
                <span class="text-hint">Feels Like</span>
            </div>
            <div class="weather__measurements__measurement">
                <div class="weather__measurements__humidity">
                    {data.humidity}%
                </div>
                <span class="text-hint">Humidity</span>
            </div>
        </div>
    </div>
</Box>

<style lang="scss">
    .weather {
        padding: 0.5rem 0.75rem;

        &__top {
            display: flex;
            align-items: center;
            gap: 1.5rem;
            height: 4rem;

            &__icon {
                font-size: 4rem;
            }

            &__labels {
                &__title {
                    font-size: 1.125rem;
                    font-weight: bold;
                }
                &__description {
                    font-size: 0.9rem;
                    color: var(--clr-text-hint);
                }
            }
        }
        &__measurements {
            margin-top: 1rem;
            display: flex;
            gap: 2rem;

            &__measurement {
                font-size: 1.1rem;

                .text-hint {
                    font-size: 0.8rem;
                }
            }
        }
    }
</style>
