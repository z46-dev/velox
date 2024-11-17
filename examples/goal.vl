// Single line comment
/*
Multi line
comment
******
**
 */

#define FOO 1
#define BAR 2.5

int add(int a, int b) {
    return a + b;
}

void noPointers(int x) {
    x *= 2;
}

class MyClass {
    int x = 0, y = 5;
    int #privateVariable = 0;

    New(int x, int y) {
        this.x = x;
        this.y = y;

        this.#privateVariable = x * y;
    }

    int getProduct() {
        return this.#privateVariable;
    }
};

int main() {
    int x = 5;

    printf(x); // 5

    x += 5;

    printf(x); // 10
    printf(x::prev); // 5

    noPointers(x);
    printf(x); // 20

    noPointers(&x);
    printf(x); // 20

    int y = 10;
    while (y < 20) {
        printf(y ++);
    }

    int list[] = {1, 2, 3, 4, 5};

    y = 0;
    while (y < 5) {
        list[y] *= 2;
        printf(list[y ++]);
    }

    MyClass myClass = MyClass(5, 10);
    printf(myClass.getProduct()); // 50
}