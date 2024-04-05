#include <SDL/SDL.h>

const int SCREEN_WIDTH = 640;
const int SCREEN_HEIGHT = 480;
const int BITS_PER_PIXEL = 32;

const SDLKey BTN_A = SDLK_SPACE;
const SDLKey BTN_MENU = SDLK_ESCAPE;

SDL_Surface *video;
SDL_Surface *screen;
SDL_Surface *background;

int pollEvents() {
    SDL_Event event;
    while (SDL_PollEvent(&event)) {
        if (event.type == SDL_QUIT) {
            return 0;
            continue;
        }

        if (event.type == SDL_KEYUP) {
            switch (event.key.keysym.sym) {
                case SDLK_SPACE:
                    return 1;
                    break;
                case SDLK_ESCAPE:
                    return 0;
                    continue;
            }
        }
    }

    return -1;
};

int refreshScreenPtr(unsigned char *pixels) {
    background = SDL_CreateRGBSurfaceFrom(
            pixels,
            SCREEN_WIDTH,
            SCREEN_HEIGHT,
            BITS_PER_PIXEL,
            SCREEN_WIDTH * 4,
            0x000000ff, 0x0000ff00, 0x00ff0000, 0xff000000
    );

    SDL_BlitSurface(background, NULL, screen, NULL);
    SDL_BlitSurface(screen, NULL, video, NULL);
    SDL_Flip(video);
   
    return 0;
}

int init() {
    SDL_Init(SDL_INIT_VIDEO);

    video = SDL_SetVideoMode(
            SCREEN_WIDTH,
            SCREEN_HEIGHT,
            BITS_PER_PIXEL,
            SDL_HWSURFACE | SDL_DOUBLEBUF
    );

    screen = SDL_CreateRGBSurface(
            SDL_HWSURFACE,
            SCREEN_WIDTH,
            SCREEN_HEIGHT,
            BITS_PER_PIXEL,
            0, 0, 0, 0
    );

    return 0;
}

void quit() {
    SDL_FreeSurface(background);
    SDL_FreeSurface(screen);
    SDL_FreeSurface(video);
    SDL_Quit();
}