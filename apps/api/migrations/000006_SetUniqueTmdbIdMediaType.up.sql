alter table public.user_medias
    drop constraint user_medias_media_id_fkey;

alter table public.medias
    drop constraint medias_tmdb_id_key;

update public.user_medias um set media_id = m.id
from public.medias m
where m.tmdb_id = um.media_id;

alter table public.medias
    add constraint medias_tmdb_id_media_type_key unique (tmdb_id, media_type);

alter table public.user_medias
    add constraint user_medias_medias_id_fkey
        foreign key (media_id) references public.medias;