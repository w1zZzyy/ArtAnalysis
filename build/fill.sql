INSERT INTO public.users (login, password, is_moderator) VALUES
('user1', 'password1', false),
('moderator1', 'modpassword', true);

INSERT INTO public.art_experts (id_artcenter, title, description, status, img_url, algorithm) VALUES
(1, 'Анализ композиционного центра картины', 
    'Определение ключевой точки композиции, выявление фокуса и направления взгляда.', 
    true, 
    'http://localhost:9000/art-center/abstract_1.jpg', 
    'Визуальный анализ изображения'
),
(2, 'Цветовой анализ произведений', 
    'Определение доминирующих цветов, контрастов и гармонии в картине или фотографии.', 
    true, 
    'http://localhost:9000/art-center/abstract_2.jpg', 
    'Анализ цветовой гармонии изображения'
),
(3, 'Оценка композиции фотографий', 
    'Выявление сильных и слабых сторон композиции фотографии, рекомендации по улучшению.', 
    true, 
    'http://localhost:9000/art-center/abstract_3.jpg', 
    'Цифровой анализ'
),
(4, 'Анализ скульптур и объектов', 
    'Определение композиционного центра и перспективного восприятия объема объекта.', 
    true, 
    'http://localhost:9000/art-center/abstract_4.jpg', 
    '3D визуализация'
),
(5, 'Композиционный анализ иллюстраций', 
    'Определение ключевых элементов иллюстрации и построение визуального фокуса.', 
    true, 
    'http://localhost:9000/art-center/abstract_5.jpg', 
    'Визуальный и цифровой анализ');

INSERT INTO public.analysis_orders (
    id_creator,
    order_status,
    date_created,
    result_x,
    result_y
) VALUES (
    1,
    'черновик',
    NOW(), 
    88, 
    92
);

INSERT INTO public.experts_to_orders (id_artcenter, id_order) VALUES
(1, 1),
(2, 1),
(3, 1);
