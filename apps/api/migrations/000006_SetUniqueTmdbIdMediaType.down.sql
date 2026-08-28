alter table public.user_medias
    drop constraint user_medias_medias_id_fkey;
    
alter table public.medias
    drop constraint medias_tmdb_id_media_type_key;

update public.user_medias um set media_id = m.tmdb_id
from public.medias m
where m.id = um.media_id;

alter table public.medias
    add constraint medias_tmdb_id_key unique (tmdb_id);

alter table public.user_medias
    add constraint user_medias_media_id_fkey
        foreign key (media_id) references public.medias (tmdb_id);