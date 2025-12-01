INSERT INTO public.users (login, password, is_moderator) VALUES
('user1', '$2a$10$NYcHfi5m9PK1wcop1Dnu.eqxIGvLJbNigg1BQwjW3UkR4vfvdN22i', false), -- password1
('moderator1', '$2a$10$NnrzLn4eE5d2sLLGP1PG4.FZNaTsvH//nV/BiP6XEgSLxjEmJiS8a', true); -- modpassword 


INSERT INTO public.art_experts (id_artcenter, title, description, status, img_url, algorithm, name) VALUES
(1, 'Анализ композиционного центра картины', 
    'Определение ключевой точки композиции, выявление фокуса и направления взгляда.', 
    true, 
    'http://localhost:9000/art-center/abstract_1.jpg', 
    'Визуальный анализ изображения',
    'Петр Иванов'
),
(2, 'Цветовой анализ произведений', 
    'Комплексное исследование цветовой структуры художественных произведений. Анализ выявляет доминирующие цветовые палитры, контрасты и гармонические сочетания. 
Включает определение основных цветовых схем, распределение теплых и холодных тонов, оценку визуального воздействия цветовых комбинаций. Позволяет раскрыть художественный замысел через анализ цветовой выразительности.', 
    true, 
    'http://localhost:9000/art-center/abstract_2.jpg', 
    'Анализ цветовой гармонии изображения', 
    'Анна Смирнова'
),
(3, 'Оценка композиции фотографий', 
    'Выявление сильных и слабых сторон композиции фотографии, рекомендации по улучшению.', 
    true, 
    'http://localhost:9000/art-center/abstract_3.jpg', 
    'Цифровой анализ',
    'Иван Петров'
),
(4, 'Анализ скульптур и объектов', 
    'Определение композиционного центра и перспективного восприятия объема объекта.', 
    true, 
    'http://localhost:9000/art-center/abstract_4.jpg', 
    '3D визуализация', 
    'Василий Березов'
),
(5, 'Композиционный анализ иллюстраций', 
    'Определение ключевых элементов иллюстрации и построение визуального фокуса.', 
    true, 
    'http://localhost:9000/art-center/abstract_5.jpg', 
    'Визуальный и цифровой анализ', 
    'Глеб Орлов'
);

INSERT INTO public.center_requests (
    id_request,
    request_status,
    date_created,
    id_creator,
    date_conclusion,
    id_moderator,
    request_description,
    factor_x,
    factor_y
) VALUES (
    1,
    'черновик',
    NOW(), 
    1,
    NULL,
    NULL,
    'тестовое описание',
    0.1241, 
    1.9801
);

INSERT INTO public.experts_to_requests (id_artcenter, id_request, center_x, center_y) VALUES
(1, 1, 10, 22),
(2, 1, 19, 72),
(3, 1, 91, 14);
